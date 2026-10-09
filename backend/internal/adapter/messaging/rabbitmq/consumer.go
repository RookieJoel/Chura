package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/RookieJoel/Chura/backend/internal/port/in"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SprintFinishedEvent struct {
	SprintID    string   `json:"sprint_id"`
	ReporterIDs []string `json:"reporter_ids"`
}

type WorkItemEvent struct {
	EventType          string `json:"event_type"`
	WorkItemID         string `json:"work_item_id"`
	Title              string `json:"title"`
	ReporterID         string `json:"reporter_id"`
	AssigneeID         string `json:"assignee_id"`
	PreviousAssigneeID string `json:"previous_assignee_id"`
	OldStatus          string `json:"old_status"`
	NewStatus          string `json:"new_status"`
}

type Consumer struct {
	channel *amqp.Channel
	queue   string
}

func NewNotificationConsumer(connection *Connection, queue string) (*Consumer, error) {
	channel, err := connection.Channel()
	if err != nil {
		return nil, err
	}
	if err := channel.ExchangeDeclare(EventsExchange, "topic", true, false, false, false, nil); err != nil {
		_ = channel.Close()
		return nil, fmt.Errorf("declare notification exchange: %w", err)
	}
	if _, err := channel.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		_ = channel.Close()
		return nil, fmt.Errorf("declare notification queue: %w", err)
	}
	for _, routingKey := range []string{
		"sprint.finished",
		"workitem.status_changed",
		"workitem.assigned",
		"workitem.unassigned",
	} {
		if err := channel.QueueBind(queue, routingKey, EventsExchange, false, nil); err != nil {
			_ = channel.Close()
			return nil, fmt.Errorf("bind notification queue to %s: %w", routingKey, err)
		}
	}
	return &Consumer{channel: channel, queue: queue}, nil
}

func (consumer *Consumer) Consume(notifications in.NotificationService) error {
	messages, err := consumer.channel.Consume(consumer.queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume notification messages: %w", err)
	}
	for message := range messages {
		var envelope struct {
			EventType string `json:"event_type"`
		}
		if err := json.Unmarshal(message.Body, &envelope); err != nil {
			log.Printf("reject invalid notification event: %v", err)
			_ = message.Nack(false, false)
			continue
		}
		if err := handleNotificationEvent(notifications, message.Body, envelope.EventType); err != nil {
			log.Printf("retry notification event %s: %v", envelope.EventType, err)
			_ = message.Nack(false, true)
			continue
		}
		if err := message.Ack(false); err != nil {
			return fmt.Errorf("ack notification message: %w", err)
		}
	}
	return nil
}

func handleNotificationEvent(notifications in.NotificationService, payload []byte, eventType string) error {
	if eventType == "" {
		var event SprintFinishedEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			return fmt.Errorf("decode sprint.finished event: %w", err)
		}
		return notifications.SendSprintFinishedNotifications(event.SprintID, event.ReporterIDs)
	}

	var event WorkItemEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("decode %s event: %w", eventType, err)
	}
	switch eventType {
	case "workitem.status_changed":
		return sendToUniqueUsers(
			notifications,
			[]string{event.AssigneeID, event.ReporterID},
			"Work item status changed",
			fmt.Sprintf("Work item %q changed from %s to %s.", event.Title, event.OldStatus, event.NewStatus),
		)
	case "workitem.assigned":
		return sendToUniqueUsers(
			notifications,
			[]string{event.AssigneeID},
			"Work item assigned",
			fmt.Sprintf("You have been assigned to work item %q.", event.Title),
		)
	case "workitem.unassigned":
		return sendToUniqueUsers(
			notifications,
			[]string{event.PreviousAssigneeID},
			"Work item unassigned",
			fmt.Sprintf("You have been unassigned from work item %q.", event.Title),
		)
	default:
		return fmt.Errorf("unsupported notification event: %s", eventType)
	}
}

func sendToUniqueUsers(notifications in.NotificationService, userIDs []string, subject, body string) error {
	seen := make(map[string]struct{}, len(userIDs))
	for _, userID := range userIDs {
		if userID == "" {
			continue
		}
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		if err := notifications.SendWorkItemNotification(userID, subject, body); err != nil {
			return err
		}
	}
	return nil
}

func (consumer *Consumer) Close() error {
	return consumer.channel.Close()
}
