package service

import (
	"errors"
	"strings"

	"github.com/RookieJoel/Chura/backend/internal/port/out"
)

type NotificationService struct {
	recipients out.NotificationRecipientRepository
	email      out.EmailSender
}

const sprintFinishedSubject = "Sprint finished"

func NewNotificationService(
	recipients out.NotificationRecipientRepository,
	email out.EmailSender,
) *NotificationService {
	return &NotificationService{
		recipients: recipients,
		email:      email,
	}
}

func (service *NotificationService) SendToReporter(reporterID, subject, body string) error {
	return service.SendWorkItemNotification(reporterID, subject, body)
}

func (service *NotificationService) SendWorkItemNotification(userID, subject, body string) error {
	if strings.TrimSpace(userID) == "" {
		return errors.New("user_id is required")
	}
	if strings.TrimSpace(subject) == "" {
		return errors.New("subject is required")
	}
	if strings.TrimSpace(body) == "" {
		return errors.New("body is required")
	}

	email, err := service.recipients.GetEmailByUserID(userID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(email) == "" {
		return errors.New("reporter has no email")
	}

	return service.email.Send(email, subject, body)
}

func (service *NotificationService) SendSprintFinishedNotifications(
	sprintID string,
	reporterIDs []string,
) error {
	if strings.TrimSpace(sprintID) == "" {
		return errors.New("sprint_id is required")
	}

	uniqueReporterIDs := make(map[string]struct{}, len(reporterIDs))
	for _, reporterID := range reporterIDs {
		reporterID = strings.TrimSpace(reporterID)
		if reporterID != "" {
			uniqueReporterIDs[reporterID] = struct{}{}
		}
	}

	body := "Sprint " + sprintID + " has finished."
	for reporterID := range uniqueReporterIDs {
		if err := service.SendToReporter(reporterID, sprintFinishedSubject, body); err != nil {
			return err
		}
	}
	return nil
}
