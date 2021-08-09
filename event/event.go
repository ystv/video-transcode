package event

import (
	"fmt"

	"github.com/streadway/amqp"
)

type Eventer struct {
	statusQueueName string
	conn            *amqp.Connection
}

// NewEventer returns a new MQ handling object
func NewEventer(conn *amqp.Connection) (*Eventer, error) {
	p := &Eventer{conn: conn}
	ch, err := p.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	err = declareExchange(ch)
	if err != nil {
		return nil, fmt.Errorf("failed to declare encode exchange: %w", err)
	}
	p.statusQueueName, err = p.newPubSubExchange("encode-status")
	if err != nil {
		return nil, fmt.Errorf("failed to declare encode-status exchange: %w", err)
	}
	return p, nil
}

func (e *Eventer) GetChannel() (*amqp.Channel, error) {
	return e.conn.Channel()
}

func declareQueue(ch *amqp.Channel, queueName string) (amqp.Queue, error) {
	q, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return q, fmt.Errorf("failed to declare queue: %w", err)
	}
	err = ch.Qos(1, 0, false)
	if err != nil {
		return q, fmt.Errorf("failed to set Qos: %w", err)
	}
	return q, nil
}

func declareExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		"encode",           // name
		amqp.ExchangeTopic, // type
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,                // arguments
	)
}
