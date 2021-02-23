package event

import (
	"fmt"

	"github.com/streadway/amqp"
)

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
		err = fmt.Errorf("declareQueue: failed to declare queue: %w", err)
		return q, err
	}
	err = ch.Qos(1, 0, false)
	if err != nil {
		err = fmt.Errorf("declareQueue: failed to set Qos: %w", err)
		return q, err
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
