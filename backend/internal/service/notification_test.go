package service_test

import (
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/service"
)

type recipientRepository struct {
	emails map[string]string
	users  []string
}

func (repository *recipientRepository) GetEmailByUserID(userID string) (string, error) {
	repository.users = append(repository.users, userID)
	return repository.emails[userID], nil
}

type emailSender struct {
	messages []emailMessage
}

type emailMessage struct {
	to      string
	subject string
	body    string
}

func (sender *emailSender) Send(to, subject, body string) error {
	sender.messages = append(sender.messages, emailMessage{
		to:      to,
		subject: subject,
		body:    body,
	})
	return nil
}

func TestNotificationServiceSendsEmailToReporter(t *testing.T) {
	recipients := &recipientRepository{
		emails: map[string]string{"reporter-1": "reporter@example.com"},
	}
	sender := &emailSender{}
	notificationService := service.NewNotificationService(recipients, sender)

	err := notificationService.SendToReporter("reporter-1", "Work item updated", "The work item changed.")
	if err != nil {
		t.Fatalf("send notification: %v", err)
	}
	if len(recipients.users) != 1 || recipients.users[0] != "reporter-1" {
		t.Fatalf("expected reporter-1 lookup, got %v", recipients.users)
	}
	if len(sender.messages) != 1 {
		t.Fatalf("expected one email, got %d", len(sender.messages))
	}
	message := sender.messages[0]
	if message.to != "reporter@example.com" {
		t.Fatalf("expected reporter email, got %q", message.to)
	}
	if message.subject != "Work item updated" || message.body != "The work item changed." {
		t.Fatalf("unexpected email: %#v", message)
	}
}

func TestNotificationServiceRequiresReporterID(t *testing.T) {
	notificationService := service.NewNotificationService(
		&recipientRepository{
			emails: map[string]string{"reporter-1": "reporter@example.com"},
		},
		&emailSender{},
	)

	if err := notificationService.SendToReporter(" ", "Subject", "Body"); err == nil {
		t.Fatal("expected reporter_id validation error")
	}
}

func TestNotificationServiceSendsOneEmailPerReporterForSprint(t *testing.T) {
	recipients := &recipientRepository{
		emails: map[string]string{
			"reporter-1": "one@example.com",
			"reporter-2": "two@example.com",
		},
	}
	sender := &emailSender{}
	notificationService := service.NewNotificationService(recipients, sender)

	err := notificationService.SendSprintFinishedNotifications(
		"sprint-1",
		[]string{"reporter-1", "reporter-1", "reporter-2", " ", "reporter-2"},
	)
	if err != nil {
		t.Fatalf("send sprint notifications: %v", err)
	}
	if len(sender.messages) != 2 {
		t.Fatalf("expected one email per unique reporter, got %d", len(sender.messages))
	}

	received := make(map[string]emailMessage, len(sender.messages))
	for _, message := range sender.messages {
		received[message.to] = message
	}
	for _, email := range []string{"one@example.com", "two@example.com"} {
		message, ok := received[email]
		if !ok {
			t.Fatalf("expected email for %s, got %v", email, received)
		}
		if message.subject != "Sprint finished" || message.body != "Sprint sprint-1 has finished." {
			t.Fatalf("unexpected sprint email: %#v", message)
		}
	}
}
