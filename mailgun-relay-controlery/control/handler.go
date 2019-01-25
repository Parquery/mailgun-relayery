package control

// Automatically generated file by swagger_to. DO NOT EDIT OR APPEND ANYTHING!

import "net/http"

// Handler defines an interface to handling the routes.
type Handler interface {
	// PutChannel handles the path `/api/channel` with the method "put".
	//
	// Path description:
	// Updates the channel uniquely identified by a descriptor.
	//
	// If there is already a channel associated with the descriptor, the old channel is overwritten with the new one.
	//
	// In order to enforce the min_period between messages, the Relay server keeps track of the time of the most
	// recently relayed message for each descriptor. If a channel is overwritten, the time of relay of the most
	// recent message is erased unless the new channel has the same min_period field as the old one.
	PutChannel(w http.ResponseWriter,
		r *http.Request,
		channel Channel)

	// DeleteChannel handles the path `/api/channel` with the method "delete".
	//
	// Path description:
	// removes the channel associated with the descriptor.
	DeleteChannel(w http.ResponseWriter,
		r *http.Request,
		descriptor Descriptor)

	// ListChannels handles the path `/api/list_channels` with the method "get".
	//
	// Path description:
	// lists the available channels information.
	ListChannels(w http.ResponseWriter,
		r *http.Request,
		page *int32,
		perPage *int32)
}

// Automatically generated file by swagger_to. DO NOT EDIT OR APPEND ANYTHING!
