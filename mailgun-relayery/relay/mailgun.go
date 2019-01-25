package relay

import (
	"fmt"

	"github.com/mailgun/mailgun-go"

	"github.com/Parquery/mailgun-relayery/mailgun-relay-controlery/control"
)

// DefaultMailgunAddress is the URL of the MailGun v3 API.
const DefaultMailgunAddress = mailgun.ApiBase

// relayMessage invokes the MailGun Go library for sending the message to the
// input channel.
//
// relayMessage requires:
// * message != nil
// * channel != nil
// * len(channel.Recipients) > 0
//
// relayMessage ensures:
// * err != nil || resp != nil
func relayMessage(message *Message,
	channel *control.Channel,
	mgData MailgunData) (resp *MailgunResponse, err error) {
	// Pre-conditions
	switch {
	case !(message != nil):
		panic("Violated: message != nil")
	case !(channel != nil):
		panic("Violated: channel != nil")
	case !(len(channel.Recipients) > 0):
		panic("Violated: len(channel.Recipients) > 0")
	default:
		// Pass
	}

	// Post-condition
	defer func() {
		if !(err != nil || resp != nil) {
			panic("Violated: err != nil || resp != nil")
		}
	}()

	// Create an instance of the MailGun Client
	mg := mailgun.NewMailgun(channel.Domain, mgData.APIKey)
	mg.SetAPIBase(mgData.Address)

	email := mg.NewMessage(entityToMailgunEmail(channel.Sender),
		message.Subject,
		message.Content)
	for _, recipient := range channel.Recipients {
		err = email.AddRecipient(entityToMailgunEmail(recipient))
		if err != nil {
			err = fmt.Errorf("error while adding recipient %#v: %s",
				recipient, err.Error())
			return
		}
	}
	for _, cc := range channel.Cc {
		email.AddCC(entityToMailgunEmail(cc))
	}
	for _, bcc := range channel.Bcc {
		email.AddBCC(entityToMailgunEmail(bcc))
	}

	if message.Html != nil {
		email.SetHtml("<html>" + *message.Html + "</html>")
	}

	human, msgID, err := mg.Send(email)
	if err != nil {
		err = fmt.Errorf("error while sending the message: %s",
			err.Error())
		return
	}

	resp = &MailgunResponse{Human: human, MsgID: msgID}
	return
}

func entityToMailgunEmail(entity control.Entity) string {
	if entity.Name != nil {
		return *entity.Name + " <" + entity.Email + ">"
	}
	return entity.Email
}

// MailgunData holds the API key and the server address of MailGun.
type MailgunData struct {
	APIKey  string
	Address string
}

// MailgunResponse the response of the MailGun API call.
type MailgunResponse struct {
	Human string
	MsgID string
}
