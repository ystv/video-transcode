package event

import (
	"fmt"

	"github.com/streadway/amqp"
)

// TranscodeVODTask represents a task to transcode for VOD
type TranscodeVODTask struct {
	Src        string `json:"src"`        // Location of source file on CDN
	Dst        string `json:"dst"`        // Destination of finished encode on CDN
	EncodeName string `json:"encodeName"` // Here for pretty logging
	EncodeArgs string `json:"encodeArgs"` // Encode arguments
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
