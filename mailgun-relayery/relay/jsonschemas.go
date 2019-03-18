package relay

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

var jsonSchemaMessageText = `{
  "title": "Message",
  "$schema": "http://json-schema.org/draft-04/schema#",
  "definitions": {
    "Message": {
      "description": "represents a message to be relayed.",
      "type": "object",
      "properties": {
        "subject": {
          "description": "contains the text to be used as the email's subject.",
          "type": "string",
          "example": "broken pipeline observed"
        },
        "content": {
          "description": "contains the text to be used as the email's content.",
          "type": "string",
          "example": "A broken pipeline was observed the 10/12/2018 at 14:37. Please contact the system operator."
        },
        "html": {
          "description": "contains the optional html text to be used as the email's content.\n\nIf set, the \"content\" field of the Message is ignored.",
          "type": "string",
          "example": "A <b>broken</b> pipeline was observed the 10/12/2018 at 14:37. Please contact the system operator."
        }
      },
      "required": [
        "subject",
        "content"
      ]
    }
  },
  "$ref": "#/definitions/Message"
}`

var jsonSchemaTokenText = `{
  "title": "Token",
  "$schema": "http://json-schema.org/draft-04/schema#",
  "description": "is a string authenticating the sender of an HTTP request.",
  "type": "string",
  "example": "RBbPYhmPXurT8nM5TAJpPOcHMaFkJblA62mr6MCvpF4oVa6cy"
}`

var jsonSchemaDescriptorText = `{
  "title": "Descriptor",
  "$schema": "http://json-schema.org/draft-04/schema#",
  "description": "identifies a channel.",
  "type": "string",
  "example": "client-1/pipeline-3"
}`

var jsonSchemaMessage = mustNewJSONSchema(
	jsonSchemaMessageText,
	"Message")

var jsonSchemaToken = mustNewJSONSchema(
	jsonSchemaTokenText,
	"Token")

var jsonSchemaDescriptor = mustNewJSONSchema(
	jsonSchemaDescriptorText,
	"Descriptor")

// ValidateAgainstMessageSchema validates a message coming from the client against Message schema.
func ValidateAgainstMessageSchema(bb []byte) error {
	loader := gojsonschema.NewStringLoader(string(bb))
	result, err := jsonSchemaMessage.Validate(loader)
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

// Automatically generated file by swagger_to. DO NOT EDIT OR APPEND ANYTHING!
