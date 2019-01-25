package control

// Automatically generated file by swagger_to. DO NOT EDIT OR APPEND ANYTHING!

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
)

// SetupRouter sets up a router. If you don't use any middleware, you are good to go.
// Otherwise, you need to maually re-implement this function with your middlewares.
func SetupRouter(h Handler) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc(`/api/channel`,
		func(w http.ResponseWriter, r *http.Request) {
			WrapPutChannel(h, w, r)
		}).Methods("put")

	r.HandleFunc(`/api/channel`,
		func(w http.ResponseWriter, r *http.Request) {
			WrapDeleteChannel(h, w, r)
		}).Methods("delete")

	r.HandleFunc(`/api/list_channels`,
		func(w http.ResponseWriter, r *http.Request) {
			WrapListChannels(h, w, r)
		}).Methods("get")

	return r
}

// WrapPutChannel wraps the path `/api/channel` with the method "put"
//
// Path description:
// Updates the channel uniquely identified by a descriptor.
//
// If there is already a channel associated with the descriptor, the old channel is overwritten with the new one.
//
// In order to enforce the min_period between messages, the Relay server keeps track of the time of the most
// recently relayed message for each descriptor. If a channel is overwritten, the time of relay of the most
// recent message is erased unless the new channel has the same min_period field as the old one.
func WrapPutChannel(h Handler, w http.ResponseWriter, r *http.Request) {
	var aChannel Channel

	if r.Body == nil {
		http.Error(w, "Parameter 'channel' expected in body, but got no body", http.StatusBadRequest)
		return
	}
	{
		var err error
		r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Body unreadable: "+err.Error(), http.StatusBadRequest)
			return
		}

		err = ValidateAgainstChannelSchema(body)
		if err != nil {
			http.Error(w, "Failed to validate against schema: "+err.Error(), http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(body, &aChannel)
		if err != nil {
			http.Error(w, "Error JSON-decoding body parameter 'channel': "+err.Error(),
				http.StatusBadRequest)
			return
		}
	}

	h.PutChannel(w,
		r,
		aChannel)
}

// WrapDeleteChannel wraps the path `/api/channel` with the method "delete"
//
// Path description:
// removes the channel associated with the descriptor.
func WrapDeleteChannel(h Handler, w http.ResponseWriter, r *http.Request) {
	var aDescriptor Descriptor

	if r.Body == nil {
		http.Error(w, "Parameter 'descriptor' expected in body, but got no body", http.StatusBadRequest)
		return
	}
	{
		var err error
		r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Body unreadable: "+err.Error(), http.StatusBadRequest)
			return
		}

		err = ValidateAgainstDescriptorSchema(body)
		if err != nil {
			http.Error(w, "Failed to validate against schema: "+err.Error(), http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(body, &aDescriptor)
		if err != nil {
			http.Error(w, "Error JSON-decoding body parameter 'descriptor': "+err.Error(),
				http.StatusBadRequest)
			return
		}
	}

	h.DeleteChannel(w,
		r,
		aDescriptor)
}

// WrapListChannels wraps the path `/api/list_channels` with the method "get"
//
// Path description:
// lists the available channels information.
func WrapListChannels(h Handler, w http.ResponseWriter, r *http.Request) {
	var aPage *int32
	var aPerPage *int32

	q := r.URL.Query()

	if _, ok := q["page"]; ok {
		{
			parsed, err := strconv.ParseInt(q.Get("page"), 10, 32)
			if err != nil {
				http.Error(w, "Parameter 'page': "+err.Error(), http.StatusBadRequest)
				return
			}
			converted := int32(parsed)
			aPage = &converted
		}
	}

	if _, ok := q["per_page"]; ok {
		{
			parsed, err := strconv.ParseInt(q.Get("per_page"), 10, 32)
			if err != nil {
				http.Error(w, "Parameter 'per_page': "+err.Error(), http.StatusBadRequest)
				return
			}
			converted := int32(parsed)
			aPerPage = &converted
		}
	}

	h.ListChannels(w,
		r,
		aPage,
		aPerPage)
}

// Automatically generated file by swagger_to. DO NOT EDIT OR APPEND ANYTHING!
