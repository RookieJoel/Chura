package rabbitmq

import (
	"fmt"
	"log"

	"github.com/RookieJoel/Chura/backend/internal/port/in"
	amqp "github.com/rabbitmq/amqp091-go"
)

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

func (consumer *Consumer) Consume(handler in.NotificationEventHandler) error {
	messages, err := consumer.channel.Consume(consumer.queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume notification messages: %w", err)
	}
	for message := range messages {
		requeue, err := handler.HandleNotificationMessage(message.Body)
		if err != nil {
			log.Printf("notification message failed: %v", err)
			_ = message.Nack(false, requeue)
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
