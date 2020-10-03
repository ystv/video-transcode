package event

import (
	"encoding/json"
	"fmt"

	"github.com/streadway/amqp"
)

// Producer for publishinh AMQP events
type Producer struct {
	conn *amqp.Connection
}

// Push (publish) a specified message to the AMQP exchange
func (e *Producer) Push(request TranscodeVODTask) error {

	reqJSON, err := json.Marshal(request)
	if err != nil {
		err = fmt.Errorf("Push: failed to marshal: %w", err)
		return err
	}

	ch, err := e.conn.Channel()
	if err != nil {
		err = fmt.Errorf("Push: failed to open channel: %w", err)
		return err
	}
	defer ch.Close()
	q, err := declareQueue(ch, "vod")
	if err != nil {
		err = fmt.Errorf("Push: failed to declare queue: %w", err)
		return err
	}
	err = ch.Publish(
		"",
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         []byte(reqJSON),
		},
	)
	if err != nil {
		err = fmt.Errorf("Push: failed to publish event \"%s\" to channel :%w", request.Src, err)
		return err
	}
	return nil
}

// NewProducer returns a new event.Producer object
// ensuring that the object is initialised, without error
func NewProducer(conn *amqp.Connection) (Producer, error) {
	p := Producer{conn: conn}
	ch, err := p.conn.Channel()
	if err != nil {
		err = fmt.Errorf("NewProducer: failed to get channel: %w", err)
		return Producer{}, err
	}
	err = declareExchange(ch)
	if err != nil {
		err = fmt.Errorf("NewProducer: failed to declare exchange: %w, err", err)
		return Producer{}, err
	}
	return p, nil
}
