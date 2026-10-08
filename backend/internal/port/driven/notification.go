package driven

type NotificationRecipientRepository interface {
	GetEmailByUserID(userID string) (string, error)
}

type EmailSender interface {
	Send(to, subject, body string) error
}
