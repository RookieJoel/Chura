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

type Consumer struct {
	channel *amqp.Channel
	queue   string
}

func NewNotificationConsumer(connection *Connection, queue string) (*Consumer, error) {
	channel, err := connection.Channel()
	if err != nil {
		return nil, err
	}
	if err := channel.ExchangeDeclare(SprintFinishedTopic, "fanout", true, false, false, false, nil); err != nil {
		_ = channel.Close()
		return nil, fmt.Errorf("declare notification exchange: %w", err)
	}
	if _, err := channel.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		_ = channel.Close()
		return nil, fmt.Errorf("declare notification queue: %w", err)
	}
	if err := channel.QueueBind(queue, "", SprintFinishedTopic, false, nil); err != nil {
		_ = channel.Close()
		return nil, fmt.Errorf("bind notification queue: %w", err)
	}
	return &Consumer{channel: channel, queue: queue}, nil
}

func (consumer *Consumer) Consume(notifications in.NotificationService) error {
	messages, err := consumer.channel.Consume(consumer.queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume notification messages: %w", err)
	}
	for message := range messages {
		var event SprintFinishedEvent
		if err := json.Unmarshal(message.Body, &event); err != nil {
			log.Printf("reject invalid sprint.finished event: %v", err)
			_ = message.Nack(false, false)
			continue
		}
		if err := notifications.SendSprintFinishedNotifications(event.SprintID, event.ReporterIDs); err != nil {
			log.Printf("retry sprint.finished notification: %v", err)
			_ = message.Nack(false, true)
			continue
		}
		if err := message.Ack(false); err != nil {
			return fmt.Errorf("ack notification message: %w", err)
		}
	}
	return nil
}

func (consumer *Consumer) Close() error {
	return consumer.channel.Close()
}
