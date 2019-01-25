package relay

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Parquery/mailgun-relayery/database"
	"github.com/Parquery/mailgun-relayery/mailgun-relay-controlery/control"
	"github.com/Parquery/mailgun-relayery/protoed"
)

// HandlerImpl implements the Handler.
type HandlerImpl struct {
	LogErr      *log.Logger
	LogOut      *log.Logger
	MailgunData MailgunData
	Env         *database.Env
}

// PutMessage implements Handler.PutMessage.
func (h *HandlerImpl) PutMessage(w http.ResponseWriter,
	r *http.Request,
	xDescriptor string,
	xToken string) {
	max := int64(1000 * 1024 * 1024) // max size of any request
	if r.ContentLength > max {
		msg := fmt.Sprintf("Request is too large (content length: %d, "+
			"max. allowed content length: %d)", r.ContentLength, max)
		http.Error(w, msg, http.StatusRequestEntityTooLarge)
		h.LogErr.Printf("error: %s: due to call to URL %s\n",
			msg, r.URL.String())
		return
	}

	var protoChan *protoed.Channel
	err := h.Env.View(func(txn *database.Txn) (txnErr error) {
		protoChan, txnErr = txn.GetChannel(xDescriptor)
		return
	})
	if err != nil {
		http.Error(w, "Failed to fetch the channel data.",
			http.StatusInternalServerError)
		h.LogErr.Printf("%s: Failed to fetch the channel data from "+
			"the database: %s\n", r.URL.String(), err.Error())
		return
	}
	if protoChan == nil {
		msg := fmt.Sprintf(
			"No channel was found for the descriptor %s", xDescriptor)
		http.Error(w, msg, http.StatusNotFound)
		h.LogErr.Printf("%s: %s\n", r.URL.String(), msg)
		return
	}

	if protoChan.Token != string(xToken) {
		msg := fmt.Sprintf("The request token for "+
			"descriptor is invalid: %s.", xDescriptor)
		http.Error(w, msg, http.StatusForbidden)
		h.LogErr.Printf("%s: %s\n", r.URL.String(), msg)
		return
	}

	chann := control.ProtoToJSON(protoChan)

	var timeLastSent *database.Timestamp
	err = h.Env.View(func(txn *database.Txn) (txnErr error) {
		timeLastSent, txnErr = txn.GetTimestamp(xDescriptor)
		return
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to check the last sending time for "+
			"descriptor %s", xDescriptor),
			http.StatusInternalServerError)
		h.LogErr.Printf("%s: Failed to fetch the data from "+
			"the timestamp database: %s\n", r.URL.String(), err.Error())
		return
	}

	if timeLastSent != nil {
		t := timeLastSent.ToTime()
		elapsedSeconds := time.Now().Sub(t).Seconds()
		if elapsedSeconds < float64(chann.MinPeriod) {
			msg := fmt.Sprintf("The minimum waiting "+
				"period of %f seconds did not elapse for descriptor %s",
				chann.MinPeriod, xDescriptor)
			http.Error(w, msg, http.StatusTooManyRequests)
			h.LogErr.Printf("%s: %s\n", r.URL.String(), msg)
			return
		}
	}

	if r.ContentLength > int64(chann.MaxSize) {
		msg := fmt.Sprintf("Request is too large. Content length is %d, "+
			"max. allowed content length is %d for descriptor %s",
			r.ContentLength, chann.MaxSize, xDescriptor)
		http.Error(w, msg, http.StatusRequestEntityTooLarge)
		h.LogErr.Printf("%s: %s\n",
			r.URL.String(), msg)
		return
	}

	// parse the form data
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "The form-data in the request is malformed",
			http.StatusBadRequest)
		h.LogErr.Printf("%s: Request form-data is malformed: got %s\n",
			r.URL.String(), err.Error())
		return
	}
	if _, ok := r.Form["message"]; !ok {
		http.Error(w, "Field 'message' expected in the form.",
			http.StatusBadRequest)
		h.LogErr.Printf("%s: failed to find the 'message' field in the form\n",
			r.URL.String())
		return
	}

	message := &Message{}
	messageStr := r.FormValue("message")
	err = ValidateAgainstMessageSchema([]byte(messageStr))

	if err != nil {
		h.LogErr.Printf("%s: Failed to validate against schema: %s\n",
			r.URL.String(), err.Error())
		http.Error(w, "Failed to validate against message schema.",
			http.StatusBadRequest)
		return
	}

	err = json.Unmarshal([]byte(messageStr), message)
	if err != nil {
		h.LogErr.Printf("%s: Failed to unmarshal the message: %s\n",
			r.URL.String(), err.Error())
		http.Error(w, "Failed to unmarshal the message.", http.StatusBadRequest)
		return
	}

	resp, err := relayMessage(message, chann, h.MailgunData)
	if err != nil {
		http.Error(w, "Failed to relay the message.",
			http.StatusInternalServerError)
		h.LogErr.Printf("%s: Failed to relay the message: %s\n",
			r.URL.String(), err.Error())
		return
	}

	ts := database.TimestampFromTime(time.Now())
	err = h.Env.Update(func(txn *database.Txn) (txnErr error) {
		txnErr = txn.PutTimestamp(database.Descriptor(xDescriptor), &ts)
		return
	})
	if err != nil {
		http.Error(w,
			"Failed to save the sending time of the message.",
			http.StatusInternalServerError)
		h.LogErr.Printf("%s: Failed to store the timestamp in the "+
			"database: %s\n", r.URL.String(), err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(
		fmt.Sprintf("The message was correctly relayed.")))
	if err != nil {
		h.LogErr.Printf("%s: Error while writing to the response "+
			"writer: %s\n", r.URL.String(), err.Error())
	}
	h.LogOut.Printf("%s: The message was correctly relayed. "+
		"Mailgun message id: %s,\n response: %s\n",
		r.URL.String(), resp.MsgID, resp.Human)
}
