package service_test

import (
	"os"
	"testing"

	emailadapter "github.com/RookieJoel/Chura/backend/internal/adapter/email"
	memory "github.com/RookieJoel/Chura/backend/internal/adapter/postgres/repository"
	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/service"
	"github.com/joho/godotenv"
)

func TestWorkItemServiceCheckSprintFinishedSendsRealEmail(t *testing.T) {
	if os.Getenv("CHURA_EMAIL_INTEGRATION") != "1" {
		t.Skip("set CHURA_EMAIL_INTEGRATION=1 to send a real email")
	}
	if err := godotenv.Load("../../.env"); err != nil {
		t.Logf("load .env: %v; using process environment", err)
	}

	required := []string{
		"SMTP_HOST",
		"SMTP_PORT",
		"SMTP_USERNAME",
		"SMTP_PASSWORD",
		"SMTP_FROM",
		"TEST_EMAIL_TO",
	}
	for _, name := range required {
		if os.Getenv(name) == "" {
			t.Fatalf("%s is required for the email integration test", name)
		}
	}

	notificationService := service.NewNotificationService(
		memory.NotificationRecipientRepository{
			TestEmail: os.Getenv("TEST_EMAIL_TO"),
		},
		emailadapter.SMTPSender{
			Host:     os.Getenv("SMTP_HOST"),
			Port:     os.Getenv("SMTP_PORT"),
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     os.Getenv("SMTP_FROM"),
		},
	)
	workItemService := service.NewWorkItemServiceWithNotifier(
		memory.NewWorkItemRepository(),
		notificationService,
	)

	for _, reporterID := range []string{"email-test-reporter-1", "email-test-reporter-2"} {
		item := validWorkItem()
		item.SprintID = "email-test-sprint"
		item.Status = domain.WorkItemStatusDone
		item.ReporterID = reporterID
		if _, err := workItemService.CreateWorkItem(item); err != nil {
			t.Fatalf("create completed work item: %v", err)
		}
	}

	finished, err := workItemService.CheckSprintFinished("email-test-sprint")
	if err != nil {
		t.Fatalf("check sprint and send email: %v", err)
	}
	if !finished {
		t.Fatal("expected sprint to be finished")
	}
}
