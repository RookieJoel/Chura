package email

import (
	"fmt"
	"net/smtp"
	"time"

	"github.com/google/uuid"
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
			"Reply-To: " + sender.From + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"Date: " + time.Now().UTC().Format(time.RFC1123Z) + "\r\n" +
			"Message-ID: <" + uuid.NewString() + "@" + sender.Host + ">\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"Content-Transfer-Encoding: 8bit\r\n" +
			"\r\n" +
			body + "\r\n",
	)
	if err := smtp.SendMail(address, auth, sender.From, []string{to}, message); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}
