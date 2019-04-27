package email

import (
	"fmt"

	mailgun "gopkg.in/mailgun/mailgun-go.v1"
)

type Client interface {
	Send(name, toAddress, subject, textBody, htmlBody string) error
}

type mgClient struct {
	from string
	mg   mailgun.Mailgun
}

func NewMailClient(domain, apiKey, publicKey string) Client {
	return &mgClient{
		from: "Support <support@lenslocked.com>",
		mg:   mailgun.NewMailgun(domain, apiKey, publicKey),
	}
}

func (mc *mgClient) Send(name, toAddress, subject, textBody, htmlBody string) error {
	msg := mc.mg.NewMessage(mc.from, subject, textBody, buildEmailField(name, toAddress))
	msg.SetHtml(htmlBody)
	_, _, err := mc.mg.Send(msg)
	return err
}

func buildEmailField(name, email string) string {
	if name == "" {
		return email
	}
	return fmt.Sprintf("%s <%s>", name, email)
}
