package relay

// Automatically generated file by swagger_to. DO NOT EDIT OR APPEND ANYTHING!

import (
	"github.com/gorilla/mux"
	"net/http"
)

// SetupRouter sets up a router. If you don't use any middleware, you are good to go.
// Otherwise, you need to maually re-implement this function with your middlewares.
func SetupRouter(h Handler) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc(`/api/message`,
		func(w http.ResponseWriter, r *http.Request) {
			WrapPutMessage(h, w, r)
		}).Methods("post")

	return r
}

// WrapPutMessage wraps the path `/api/message` with the method "post".
//
// Path description:
// sends a message to the server, which relays it to the MailGun API.
//
// The given (descriptor, token) pair are authenticated first.
// The message's metadata is determined by the channel information from the database.
func WrapPutMessage(h Handler, w http.ResponseWriter, r *http.Request) {
	var aXDescriptor string
	var aXToken string

	hdr := r.Header

	if _, ok := hdr["X-Descriptor"]; !ok {
		http.Error(w, "Parameter 'X-Descriptor' expected in header", http.StatusBadRequest)
		return
	}
	aXDescriptor = hdr.Get("X-Descriptor")

	if _, ok := hdr["X-Token"]; !ok {
		http.Error(w, "Parameter 'X-Token' expected in header", http.StatusBadRequest)
		return
	}
	aXToken = hdr.Get("X-Token")

	h.PutMessage(w,
		r,
		aXDescriptor,
		aXToken)
}

// Automatically generated file by swagger_to. DO NOT EDIT OR APPEND ANYTHING!
