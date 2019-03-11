package relay

// Automatically generated file by swagger_to. DO NOT EDIT OR APPEND ANYTHING!

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
