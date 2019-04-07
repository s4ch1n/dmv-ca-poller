package notification

import (
	"fmt"
	"log"
	"net/smtp"
	"strconv"

	"github.com/GeorgeZhai/dmv-ca-poller/outlooksmtpauth"
)

// EmailConfig contains email configuration data for SendEmail
type EmailConfig struct {
	SenderEmail   string
	SenderPW      string
	ReceiverEmail string
	ServerHost    string
	ServerPort    int
}

// SendEmail use EmailConfig to send string content
func SendEmail(c EmailConfig, s string) {
	log.Printf("sending email content: %s to %s", s, c.ReceiverEmail)
	auth := outlooksmtpauth.OutlookSmtpAuth(c.SenderEmail, c.SenderPW)

	to := []string{c.ReceiverEmail}
	msgs := fmt.Sprintf("To: %s\r\n"+
		"Subject: DMV search notification\r\n"+
		"\r\n"+
		" %s \r\n", c.ReceiverEmail, s)
	msg := []byte(msgs)
	err := smtp.SendMail(c.ServerHost+":"+strconv.Itoa(c.ServerPort), auth, c.SenderEmail, to, msg)
	if err != nil {
		log.Fatal(err)
	}

}
