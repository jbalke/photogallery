package email

import (
	"fmt"
	"net/url"

	mailgun "gopkg.in/mailgun/mailgun-go.v1"
)

const (
	siteURL      = "https://wwww.lenslocked.com"
	resetBaseURL = siteURL + "/reset"

	welcomeTextBody = `Hi %s!
	
	Welcome to Lenslocked.com, we hope you enjoy our site!
	
	Best Wishes
	John`

	welcomeHTMLBody = `<p>Hi %s!</p>
	<p>Welcome to <a href="https://www.lenslocked.com">Lenslocked.com</a>!</p>
	<p>Best Wishes<br>John</p>`

	welcomeSubject = "Welcome to Lenslocked.com!"
	resetSubject   = "Reset password instructions"

	resetTextTmpl = `Hi there!

It appears that you have requested a password reset. If this was you, please follow the 
link below to update your password:

%s

Your reset token is:

%s

If you did not request a password reset you can safely ignore this email, your account will not change.

Best,
LensLocked Support
`

	resetHTMLTmpl = `<p>Hi there!</p>

<p>It appears that you have requested a password reset. If this was you, please follow the <br>
link below to update your password:</p>

<a href="%s">%s</a><br>

<p>Your reset token is:</p>

%s<br>

<p>If you did not request a password reset you can safely ignore this email,<br>
your account will not change.</p>

<p>Best,<br>
LensLocked Support</p>
`
)

type MailClient interface {
	Send(name, toAddress, subject, textBody, htmlBody string) error
	Welcome(name, toAddress string) error
	ResetPw(toAddress, token string) error
}

type Client struct {
	from string
	mg   mailgun.Mailgun
}

type MailConfig func(*Client)

func WithSender(name, email string) MailConfig {
	return func(c *Client) {
		c.from = buildEmailField(name, email)
	}
}

func WithMailgun(domain, apiKey, publicKey string) MailConfig {
	return func(c *Client) {
		c.mg = mailgun.NewMailgun(domain, apiKey, publicKey)
	}
}

func NewClient(cfgs ...MailConfig) MailClient {
	client := Client{
		from: "support@lenslocked.com",
	}
	for _, cfg := range cfgs {
		cfg(&client)
	}
	return &client
}

func (mc *Client) Send(name, toAddress, subject, textBody, htmlBody string) error {
	msg := mc.mg.NewMessage(mc.from, subject, textBody, buildEmailField(name, toAddress))
	msg.SetHtml(htmlBody)
	_, _, err := mc.mg.Send(msg)
	return err
}

func (mc *Client) Welcome(name, toAddress string) error {
	emailText := fmt.Sprintf(welcomeTextBody, name)
	emailHTML := fmt.Sprintf(welcomeHTMLBody, name)
	msg := mc.mg.NewMessage(mc.from, welcomeSubject, emailText, buildEmailField(name, toAddress))
	msg.SetHtml(emailHTML)
	_, _, err := mc.mg.Send(msg)
	return err
}

func (mc *Client) ResetPw(toAddress, token string) error {
	v := url.Values{}
	v.Set("token", token)
	resetURL := resetBaseURL + "?" + v.Encode()
	resetText := fmt.Sprintf(resetTextTmpl, resetURL, token)
	msg := mc.mg.NewMessage(mc.from, resetSubject, resetText, toAddress)
	resetHTML := fmt.Sprintf(resetHTMLTmpl, resetURL, resetURL, token)
	msg.SetHtml(resetHTML)
	_, _, err := mc.mg.Send(msg)
	return err
}

func buildEmailField(name, email string) string {
	if name == "" {
		return email
	}
	return fmt.Sprintf("%s <%s>", name, email)
}
