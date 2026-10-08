package email

import (
	"fmt"
	"net/smtp"
)

type SMTPSender struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func (sender SMTPSender) Send(to, subject, body string) error {
	address := sender.Host + ":" + sender.Port
	auth := smtp.PlainAuth("", sender.Username, sender.Password, sender.Host)
	message := []byte(
		"From: " + sender.From + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"\r\n" +
			body + "\r\n",
	)
	if err := smtp.SendMail(address, auth, sender.From, []string{to}, message); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}
