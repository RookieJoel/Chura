package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const EventsExchange = "chura.events"
const SprintFinishedTopic = "sprint.finished"

type Publisher struct {
	channel *amqp.Channel
}

func NewPublisher(connection *Connection) (*Publisher, error) {
	channel, err := connection.Channel()
	if err != nil {
		return nil, err
	}
	if err := channel.ExchangeDeclare(EventsExchange, "topic", true, false, false, false, nil); err != nil {
		_ = channel.Close()
		return nil, fmt.Errorf("declare rabbitmq exchange: %w", err)
	}
	return &Publisher{channel: channel}, nil
}

func (publisher *Publisher) Publish(topic string, payload []byte) error {
	if err := publisher.channel.Publish(EventsExchange, topic, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        payload,
	}); err != nil {
		return fmt.Errorf("publish %s event: %w", topic, err)
	}
	return nil
}

func (publisher *Publisher) Close() error {
	return publisher.channel.Close()
}
