package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Connection struct {
	connection *amqp.Connection
}

func Connect(url string) (*Connection, error) {
	if url == "" {
		return nil, fmt.Errorf("rabbitmq URL is required")
	}
	connection, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connect to rabbitmq: %w", err)
	}
	return &Connection{connection: connection}, nil
}

func (connection *Connection) Channel() (*amqp.Channel, error) {
	channel, err := connection.connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("open rabbitmq channel: %w", err)
	}
	return channel, nil
}

func (connection *Connection) Close() error {
	return connection.connection.Close()
}
