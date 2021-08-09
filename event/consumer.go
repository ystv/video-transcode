package event

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func ListenToQueues(ch *amqp.Channel, queues []string) (<-chan amqp.Delivery, error) {
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
		log.Println("listening to: " + queue)
	}
	return msgChan, nil
}
