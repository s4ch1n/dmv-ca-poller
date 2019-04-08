package notification

import (
	"fmt"
	"log"
	"net/smtp"
	// "github.com/GeorgeZhai/dmv-ca-poller/outlooksmtpauth"
)

// EmailClient contains email configuration data for SendEmail
type EmailClient struct {
	SenderEmail   string
	SenderPW      string
	ReceiverEmail string
	ServerHost    string
	ServerPort    int
}

// NewEmailClient create a new EmailClient
func NewEmailClient(SenderEmail string, SenderPW string, ReceiverEmail string, ServerHost string, ServerPort int) *EmailClient {
	return &EmailClient{
		SenderEmail:   SenderEmail,
		SenderPW:      SenderPW,
		ReceiverEmail: ReceiverEmail,
		ServerHost:    ServerHost,
		ServerPort:    ServerPort,
	}
}

// SendEmail use EmailConfig to send string content
func (c *EmailClient) SendEmail(s string) error {
	log.Printf("sending email content: %s to %s", s, c.ReceiverEmail)
	auth := OutlookSmtpAuth(c.SenderEmail, c.SenderPW)

	to := []string{c.ReceiverEmail}
	msgs := fmt.Sprintf("To: %s\r\n"+
		"Subject: DMV search notification\r\n"+
		"\r\n"+
		" %s \r\n", c.ReceiverEmail, s)
	msg := []byte(msgs)
	return smtp.SendMail(fmt.Sprintf("%s:%d", c.ServerHost, c.ServerPort), auth, c.SenderEmail, to, msg)
}
