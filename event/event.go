package event

import (
	"fmt"

	"github.com/streadway/amqp"
)

const (
	queueVideoSimple string = "video/simple"
	queueVideoVOD    string = "video/vod"
	queueImageSimple string = "image/simple"
)

func newListener(ch *amqp.Channel, queues []string) (<-chan amqp.Delivery, error) {
	msgChan := make(<-chan amqp.Delivery)
	for _, queue := range queues {
		q, err := declareQueue(ch, queue)
		if err != nil {
			return nil, fmt.Errorf("failed to declare queue \"%s\"", queue)
		}
		msgChan, err = ch.Consume(
			q.Name, // queue
			"",     // consumer
			false,  // autoAck
			false,  // exclusive
			false,  // noLocal
			false,  // noWait
			nil,    // args
		)
		if err != nil {
			return nil, fmt.Errorf("Listen: failed to consume queue: %w", err)
		}
	}
	return msgChan, nil
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
