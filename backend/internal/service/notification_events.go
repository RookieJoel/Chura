package service

import (
	"encoding/json"
	"fmt"

	"github.com/RookieJoel/Chura/backend/internal/port/in"
)

type NotificationEventHandler struct {
	notifications in.NotificationService
}

var _ in.NotificationEventHandler = (*NotificationEventHandler)(nil)

type notificationEvent struct {
	EventType          string   `json:"event_type"`
	SprintID           string   `json:"sprint_id"`
	ReporterIDs        []string `json:"reporter_ids"`
	Title              string   `json:"title"`
	ReporterID         string   `json:"reporter_id"`
	AssigneeID         string   `json:"assignee_id"`
	PreviousAssigneeID string   `json:"previous_assignee_id"`
	OldStatus          string   `json:"old_status"`
	NewStatus          string   `json:"new_status"`
}

func NewNotificationEventHandler(notifications in.NotificationService) *NotificationEventHandler {
	return &NotificationEventHandler{notifications: notifications}
}

func (handler *NotificationEventHandler) HandleNotificationMessage(payload []byte) (bool, error) {
	var event notificationEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return false, fmt.Errorf("decode notification event: %w", err)
	}
	if err := handler.handle(event); err != nil {
		return true, err
	}
	return false, nil
}

func (handler *NotificationEventHandler) handle(event notificationEvent) error {
	if event.EventType == "" {
		return handler.notifications.SendSprintFinishedNotifications(event.SprintID, event.ReporterIDs)
	}

	switch event.EventType {
	case "workitem.status_changed":
		return handler.sendToUniqueUsers(
			[]string{event.AssigneeID, event.ReporterID},
			"Work item status changed",
			fmt.Sprintf("Work item %q changed from %s to %s.", event.Title, event.OldStatus, event.NewStatus),
		)
	case "workitem.assigned":
		return handler.sendToUniqueUsers(
			[]string{event.AssigneeID},
			"Work item assigned",
			fmt.Sprintf("You have been assigned to work item %q.", event.Title),
		)
	case "workitem.unassigned":
		return handler.sendToUniqueUsers(
			[]string{event.PreviousAssigneeID},
			"Work item unassigned",
			fmt.Sprintf("You have been unassigned from work item %q.", event.Title),
		)
	default:
		return fmt.Errorf("unsupported notification event: %s", event.EventType)
	}
}

func (handler *NotificationEventHandler) sendToUniqueUsers(userIDs []string, subject, body string) error {
	seen := make(map[string]struct{}, len(userIDs))
	for _, userID := range userIDs {
		if userID == "" {
			continue
		}
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		if err := handler.notifications.SendWorkItemNotification(userID, subject, body); err != nil {
			return err
		}
	}
	return nil
}
