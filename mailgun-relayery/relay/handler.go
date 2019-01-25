package relay

// Automatically generated file by swagger_to. DO NOT EDIT OR APPEND ANYTHING!

import "net/http"

// Handler defines an interface to handling the routes.
type Handler interface {
	// PutMessage handles the path `/api/message` with the method "post".
	//
	// Path description:
	// sends a message to the server, which relays it to the MailGun API.
	//
	// The given (descriptor, token) pair are authenticated first.
	// The message's metadata is determined by the channel information from the database.
	PutMessage(w http.ResponseWriter,
		r *http.Request,
		xDescriptor string,
		xToken string)
}

// Automatically generated file by swagger_to. DO NOT EDIT OR APPEND ANYTHING!
