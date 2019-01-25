package control

// Automatically generated file by swagger_to. DO NOT EDIT OR APPEND ANYTHING!

// Token is a string authenticating the sender of an HTTP request.
type Token string

// Descriptor identifies a channel.
type Descriptor string

// Entity contains the email address and optionally the name of an entity.
type Entity struct {
	Email string `json:"email"`

	Name *string `json:"name,omitempty"`
}

// Channel defines the messaging channel.
type Channel struct {
	Descriptor Descriptor `json:"descriptor"`

	Token Token `json:"token"`

	Sender Entity `json:"sender"`

	Recipients []Entity `json:"recipients"`

	Cc []Entity `json:"cc,omitempty"`

	Bcc []Entity `json:"bcc,omitempty"`

	// indicates the MailGun domain for the channel.
	Domain string `json:"domain"`

	// is the minimum push period frequency for a channel, in seconds.
	MinPeriod float32 `json:"min_period"`

	// indicates the maximum allowed size of the request, in bytes.
	MaxSize int32 `json:"max_size"`
}

// ChannelsPage lists channels in a paginated manner.
type ChannelsPage struct {
	// specifies the index of the page.
	Page int32 `json:"page"`

	// specifies the number of pages available.
	PageCount int32 `json:"page_count"`

	// specifies the number of items per page.
	PerPage int32 `json:"per_page"`

	// contains the channel data.
	Channels []Channel `json:"channels"`
}
