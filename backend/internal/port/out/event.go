package out

type EventPublisher interface {
	Publish(topic string, payload []byte) error
}
