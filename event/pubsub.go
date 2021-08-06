package event

import (
	"fmt"

	"github.com/streadway/amqp"
)

func (e *Eventer) newPubSubExchange(exchangeName string) (string, error) {
	ch, err := e.conn.Channel()
	if err != nil {
		return "", fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchangeName,        // name
		amqp.ExchangeFanout, // type
		true,                // durable
		false,               // auto-delete
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return "", fmt.Errorf("failed to declare exchange: %w", err)
	}

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)

	return q.Name, nil
}

func (e *Eventer) SendStatus(exchangeName string, reqJSON []byte) error {
	ch, err := e.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	err = ch.QueueBind(
		e.statusQueueName, // queue name
		"",                // routing key
		"encode-status",   // exchange
		false,             // no-wait
		nil,               // arguments
	)

	err = ch.Publish(
		exchangeName, // exchange
		"",           // key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(reqJSON),
		})
	if err != nil {
		return fmt.Errorf("failed to send message to exchange: %w", err)
	}
	return nil
}
