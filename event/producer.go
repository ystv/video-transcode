package event

import (
	"encoding/json"
	"fmt"

	"github.com/streadway/amqp"
	"github.com/ystv/video-transcode/task"
)

// Push (publish) a specified message to the AMQP exchange
func (e *Eventer) Push(request task.Task, taskType string) error {
	reqJSON, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	ch, err := e.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()
	q, err := declareQueue(ch, taskType)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
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
		return fmt.Errorf("Push: failed to publish event \"%s\" to channel :%w", request.GetID(), err)
	}
	return nil
}
