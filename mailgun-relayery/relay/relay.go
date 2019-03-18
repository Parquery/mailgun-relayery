package relay

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/Parquery/mailgun-relayery/database"
	"github.com/Parquery/mailgun-relayery/mailgun-relay-controlery/control"
	"github.com/Parquery/mailgun-relayery/protoed"
)

// Token is a string authenticating the sender of an HTTP request.
type Token string

// Descriptor identifies a channel.
type Descriptor string

// Message represents a message to be relayed.
type Message struct {
	// contains the text to be used as the email's subject.
	Subject string `json:"subject"`

	// contains the text to be used as the email's content.
	Content string `json:"content"`

	// contains the optional html text to be used as the email's content.
	//
	// If set, the "content" field of the Message is ignored.
	HTML *string `json:"html,omitempty"`
}

// Handler holds the global dependencies for handling the routes.
type Handler struct {
	LogErr      *log.Logger
	LogOut      *log.Logger
	MailgunData MailgunData
	Env         *database.Env
}

// SetupRouter sets up a router. If you don't use any middleware, you are good to go.
// Otherwise, you need to maually re-implement this function with your middlewares.
func SetupRouter(h *Handler) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc(`/api/message`,
		func(w http.ResponseWriter, r *http.Request) {
			PutMessage(h, w, r)
		}).Methods("post")

	return r
}

// PutMessage sends a message to the server, which relays it to the MailGun API.
//
// The given (descriptor, token) pair are authenticated first.
// The message's metadata is determined by the channel information from the database.
func PutMessage(h *Handler, w http.ResponseWriter, r *http.Request) {
	var xDescriptor string
	var xToken string

	////
	// Parse the header
	////

	hdr := r.Header

	if _, ok := hdr["X-Descriptor"]; !ok {
		http.Error(w, "Parameter 'X-Descriptor' expected in header", http.StatusBadRequest)
		return
	}
	xDescriptor = hdr.Get("X-Descriptor")

	if _, ok := hdr["X-Token"]; !ok {
		http.Error(w, "Parameter 'X-Token' expected in header", http.StatusBadRequest)
		return
	}
	xToken = hdr.Get("X-Token")

	if r.Body == nil {
		http.Error(w, "Parameter 'message' expected in body, but got no body", http.StatusBadRequest)
		return
	}

	////
	// Get channel information
	////

	var protoChan *protoed.Channel
	err := h.Env.View(func(txn *database.Txn) (txnErr error) {
		protoChan, txnErr = txn.GetChannel(xDescriptor)
		return
	})
	if err != nil {
		http.Error(w, "Failed to fetch the channel data from the database.",
			http.StatusInternalServerError)
		h.LogErr.Printf("%s: Failed to fetch the channel data from "+
			"the database: %s\n", r.URL.String(), err.Error())
		return
	}

	////
	// Verify the (descriptor, token) pair
	////

	if protoChan == nil {
		msg := fmt.Sprintf(
			"No channel was found for the descriptor: %s", xDescriptor)
		http.Error(w, msg, http.StatusNotFound)
		h.LogErr.Printf("%s: %s\n", r.URL.String(), msg)
		return
	}

	if protoChan.Token != string(xToken) {
		msg := fmt.Sprintf("The request token for the "+
			"descriptor is invalid: %s", xDescriptor)
		http.Error(w, msg, http.StatusForbidden)
		h.LogErr.Printf("%s: %s\n", r.URL.String(), msg)
		return
	}

	chann := control.ProtoToJSON(protoChan)

	////
	// Check that this request obeys the minimum period between two requests.
	////

	var timeLastRequest *database.Timestamp
	tooSoon := false

	err = h.Env.Update(func(txn *database.Txn) (txnErr error) {
		////
		// Get
		////

		timeLastRequest, txnErr = txn.GetTimestamp(xDescriptor)
		if txnErr != nil {
			return
		}

		////
		// Check
		////

		if timeLastRequest != nil {
			t := timeLastRequest.ToTime()
			elapsedSeconds := time.Now().Sub(t).Seconds()
			if elapsedSeconds < float64(chann.MinPeriod) {
				tooSoon = true
				return
			}
		}

		////
		// Update
		////

		ts := database.TimestampFromTime(time.Now())
		txnErr = txn.PutTimestamp(database.Descriptor(xDescriptor), &ts)
		if txnErr != nil {
			return
		}

		return
	})
	if err != nil {
		http.Error(w, fmt.Sprintf(
			"Error accessing/updating the timestamp of the last request for the descriptor: %s",
			xDescriptor),
			http.StatusInternalServerError)
		h.LogErr.Printf(
			"%s: Failed to access and update the timestamp of the last request for the descriptor: %s\n",
			r.URL.String(), err.Error())
		return
	}

	if tooSoon {
		msg := fmt.Sprintf("The minimum waiting "+
			"period of %f seconds between requests "+
			"did not elapse for the descriptor: %s",
			chann.MinPeriod, xDescriptor)
		http.Error(w, msg, http.StatusTooManyRequests)
		h.LogErr.Printf("%s: %s\n", r.URL.String(), msg)
		return
	}

	////
	// Read the body
	////

	if r.ContentLength > int64(chann.MaxSize) {
		msg := fmt.Sprintf("Request is too large. Content length is %d, "+
			"max. allowed content length is %d for descriptor %s",
			r.ContentLength, chann.MaxSize, xDescriptor)
		http.Error(w, msg, http.StatusRequestEntityTooLarge)
		h.LogErr.Printf("%s: %s\n", r.URL.String(), msg)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(chann.MaxSize))
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Body unreadable: "+err.Error(), http.StatusBadRequest)
		h.LogErr.Printf("%s: body unreadable: %s\n", r.URL.String(), err.Error())
		return
	}

	////
	// Parse the message
	////

	message := &Message{}
	err = ValidateAgainstMessageSchema(body)
	if err != nil {
		h.LogErr.Printf("%s: Failed to validate against schema: %s\n",
			r.URL.String(), err.Error())
		http.Error(w, "Failed to validate against message schema.",
			http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, message)
	if err != nil {
		h.LogErr.Printf("%s: Failed to unmarshal the message: %s\n",
			r.URL.String(), err.Error())
		http.Error(w, "Failed to unmarshal the message.", http.StatusBadRequest)
		return
	}

	////
	// Relay
	////

	resp, err := relayMessage(message, chann, h.MailgunData)
	if err != nil {
		http.Error(w, "Failed to relay the message.",
			http.StatusInternalServerError)
		h.LogErr.Printf("%s: Failed to relay the message: %s\n",
			r.URL.String(), err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(
		fmt.Sprintf("The message has been correctly relayed.")))
	if err != nil {
		h.LogErr.Printf("%s: Error while writing to the response "+
			"writer: %s\n", r.URL.String(), err.Error())
	}
	h.LogOut.Printf("%s: The message has been correctly relayed. "+
		"Mailgun message id: %s,\n response: %s\n",
		r.URL.String(), resp.MsgID, resp.Human)
}
