package control

// Automatically generated file by swagger_to. DO NOT EDIT OR APPEND ANYTHING!

import (
	"errors"
	"fmt"
	"github.com/xeipuuv/gojsonschema"
)

func mustNewJSONSchema(text string, name string) *gojsonschema.Schema {
	loader := gojsonschema.NewStringLoader(text)
	schema, err := gojsonschema.NewSchema(loader)
	if err != nil {
		panic(fmt.Sprintf("failed to load JSON Schema %#v: %s", text, err.Error()))
	}
	return schema
}

var jsonSchemaChannelText = `{
  "title": "Channel",
  "$schema": "http://json-schema.org/draft-04/schema#",
  "definitions": {
    "Entity": {
      "description": "contains the email address and optionally the name of an entity.",
      "type": "object",
      "properties": {
        "email": {
          "type": "string",
          "example": "name@domain.com"
        },
        "name": {
          "type": "string",
          "example": "John Doe"
        }
      },
      "required": [
        "email"
      ]
    },
    "Token": {
      "description": "is a string authenticating the sender of an HTTP request.",
      "type": "string",
      "example": "RBbPYhmPXurT8nM5TAJpPOcHMaFkJblA62mr6MCvpF4oVa6cy"
    },
    "Descriptor": {
      "description": "identifies a channel.",
      "type": "string",
      "example": "client-1/pipeline-3"
    },
    "Channel": {
      "description": "defines the messaging channel.",
      "type": "object",
      "properties": {
        "descriptor": {
          "$ref": "#/definitions/Descriptor"
        },
        "token": {
          "$ref": "#/definitions/Token"
        },
        "sender": {
          "$ref": "#/definitions/Entity"
        },
        "recipients": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/Entity"
          }
        },
        "cc": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/Entity"
          }
        },
        "bcc": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/Entity"
          }
        },
        "domain": {
          "description": "indicates the MailGun domain for the channel.",
          "type": "string",
          "example": "marketing.domainname.com"
        },
        "min_period": {
          "description": "is the minimum push period frequency for a channel, in seconds.",
          "type": "number",
          "format": "float"
        },
        "max_size": {
          "description": "indicates the maximum allowed size of the request, in bytes.",
          "type": "integer",
          "format": "int32"
        }
      },
      "required": [
        "descriptor",
        "token",
        "sender",
        "recipients",
        "domain",
        "min_period",
        "max_size"
      ]
    }
  },
  "$ref": "#/definitions/Channel"
}`

var jsonSchemaDescriptorText = `{
  "title": "Descriptor",
  "$schema": "http://json-schema.org/draft-04/schema#",
  "definitions": {
    "Descriptor": {
      "description": "identifies a channel.",
      "type": "string",
      "example": "client-1/pipeline-3"
    }
  },
  "$ref": "#/definitions/Descriptor"
}`

var jsonSchemaTokenText = `{
  "title": "Token",
  "$schema": "http://json-schema.org/draft-04/schema#",
  "description": "is a string authenticating the sender of an HTTP request.",
  "type": "string",
  "example": "RBbPYhmPXurT8nM5TAJpPOcHMaFkJblA62mr6MCvpF4oVa6cy"
}`

var jsonSchemaEntityText = `{
  "title": "Entity",
  "$schema": "http://json-schema.org/draft-04/schema#",
  "description": "contains the email address and optionally the name of an entity.",
  "type": "object",
  "properties": {
    "email": {
      "type": "string",
      "example": "name@domain.com"
    },
    "name": {
      "type": "string",
      "example": "John Doe"
    }
  },
  "required": [
    "email"
  ]
}`

var jsonSchemaChannelsPageText = `{
  "title": "ChannelsPage",
  "$schema": "http://json-schema.org/draft-04/schema#",
  "definitions": {
    "Entity": {
      "description": "contains the email address and optionally the name of an entity.",
      "type": "object",
      "properties": {
        "email": {
          "type": "string",
          "example": "name@domain.com"
        },
        "name": {
          "type": "string",
          "example": "John Doe"
        }
      },
      "required": [
        "email"
      ]
    },
    "Token": {
      "description": "is a string authenticating the sender of an HTTP request.",
      "type": "string",
      "example": "RBbPYhmPXurT8nM5TAJpPOcHMaFkJblA62mr6MCvpF4oVa6cy"
    },
    "Descriptor": {
      "description": "identifies a channel.",
      "type": "string",
      "example": "client-1/pipeline-3"
    },
    "Channel": {
      "description": "defines the messaging channel.",
      "type": "object",
      "properties": {
        "descriptor": {
          "$ref": "#/definitions/Descriptor"
        },
        "token": {
          "$ref": "#/definitions/Token"
        },
        "sender": {
          "$ref": "#/definitions/Entity"
        },
        "recipients": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/Entity"
          }
        },
        "cc": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/Entity"
          }
        },
        "bcc": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/Entity"
          }
        },
        "domain": {
          "description": "indicates the MailGun domain for the channel.",
          "type": "string",
          "example": "marketing.domainname.com"
        },
        "min_period": {
          "description": "is the minimum push period frequency for a channel, in seconds.",
          "type": "number",
          "format": "float"
        },
        "max_size": {
          "description": "indicates the maximum allowed size of the request, in bytes.",
          "type": "integer",
          "format": "int32"
        }
      },
      "required": [
        "descriptor",
        "token",
        "sender",
        "recipients",
        "domain",
        "min_period",
        "max_size"
      ]
    }
  },
  "description": "lists channels in a paginated manner.",
  "type": "object",
  "properties": {
    "page": {
      "description": "specifies the index of the page.",
      "type": "integer",
      "format": "int32"
    },
    "page_count": {
      "description": "specifies the number of pages available.",
      "type": "integer",
      "format": "int32"
    },
    "per_page": {
      "description": "specifies the number of items per page.",
      "type": "integer",
      "format": "int32"
    },
    "channels": {
      "description": "contains the channel data.",
      "type": "array",
      "items": {
        "$ref": "#/definitions/Channel"
      }
    }
  },
  "required": [
    "page",
    "page_count",
    "per_page",
    "channels"
  ]
}`

var jsonSchemaChannel = mustNewJSONSchema(
	jsonSchemaChannelText,
	"Channel")

var jsonSchemaDescriptor = mustNewJSONSchema(
	jsonSchemaDescriptorText,
	"Descriptor")

var jsonSchemaToken = mustNewJSONSchema(
	jsonSchemaTokenText,
	"Token")

var jsonSchemaEntity = mustNewJSONSchema(
	jsonSchemaEntityText,
	"Entity")

var jsonSchemaChannelsPage = mustNewJSONSchema(
	jsonSchemaChannelsPageText,
	"ChannelsPage")

// ValidateAgainstChannelSchema validates a message coming from the client against Channel schema.
func ValidateAgainstChannelSchema(bb []byte) error {
	loader := gojsonschema.NewStringLoader(string(bb))
	result, err := jsonSchemaChannel.Validate(loader)
	if err != nil {
		return err
	}

	if result.Valid() {
		return nil
	}

	msg := ""
	for i, valErr := range result.Errors() {
		if i > 0 {
			msg += ", "
		}
		msg += valErr.String()
	}
	return errors.New(msg)
}

// ValidateAgainstDescriptorSchema validates a message coming from the client against Descriptor schema.
func ValidateAgainstDescriptorSchema(bb []byte) error {
	loader := gojsonschema.NewStringLoader(string(bb))
	result, err := jsonSchemaDescriptor.Validate(loader)
	if err != nil {
		return err
	}

	if result.Valid() {
		return nil
	}

	msg := ""
	for i, valErr := range result.Errors() {
		if i > 0 {
			msg += ", "
		}
		msg += valErr.String()
	}
	return errors.New(msg)
}

// ValidateAgainstTokenSchema validates a message coming from the client against Token schema.
func ValidateAgainstTokenSchema(bb []byte) error {
	loader := gojsonschema.NewStringLoader(string(bb))
	result, err := jsonSchemaToken.Validate(loader)
	if err != nil {
		return err
	}

	if result.Valid() {
		return nil
	}

	msg := ""
	for i, valErr := range result.Errors() {
		if i > 0 {
			msg += ", "
		}
		msg += valErr.String()
	}
	return errors.New(msg)
}

// ValidateAgainstEntitySchema validates a message coming from the client against Entity schema.
func ValidateAgainstEntitySchema(bb []byte) error {
	loader := gojsonschema.NewStringLoader(string(bb))
	result, err := jsonSchemaEntity.Validate(loader)
	if err != nil {
		return err
	}

	if result.Valid() {
		return nil
	}

	msg := ""
	for i, valErr := range result.Errors() {
		if i > 0 {
			msg += ", "
		}
		msg += valErr.String()
	}
	return errors.New(msg)
}

// ValidateAgainstChannelsPageSchema validates a message coming from the client against ChannelsPage schema.
func ValidateAgainstChannelsPageSchema(bb []byte) error {
	loader := gojsonschema.NewStringLoader(string(bb))
	result, err := jsonSchemaChannelsPage.Validate(loader)
	if err != nil {
		return err
	}

	if result.Valid() {
		return nil
	}

	msg := ""
	for i, valErr := range result.Errors() {
		if i > 0 {
			msg += ", "
		}
		msg += valErr.String()
	}
	return errors.New(msg)
}

// Automatically generated file by swagger_to. DO NOT EDIT OR APPEND ANYTHING!
