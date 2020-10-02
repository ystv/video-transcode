package event

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/streadway/amqp"
)

// Consumer for receiving AMPQ events
type Consumer struct {
	conn *amqp.Connection
}

// NewConsumer returns a new Consumer
func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	c := Consumer{conn: conn}
	ch, err := c.conn.Channel()
	if err != nil {
		err = fmt.Errorf("NewConsumer: failed to get channel: %w", err)
		return Consumer{}, err
	}
	err = declareExchange(ch)
	if err != nil {
		err = fmt.Errorf("NewConsumer: failed to declare exchange: %w, err", err)
		return Consumer{}, err
	}
	return c, nil
}

// Listen will listen for all new Queue publications
// and print them to the console.
func (c *Consumer) Listen() error {
	ch, err := c.conn.Channel()
	if err != nil {
		err = fmt.Errorf("Listen: failed to get channel: %w", err)
		return err
	}
	defer ch.Close()
	q, err := declareQueue(ch, "vod")
	if err != nil {
		err = fmt.Errorf("Listen: failed to declare queue: %w", err)
		return err
	}
	msgChan, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		err = fmt.Errorf("Listen: failed to consume queue: %w", err)
		return err
	}

	stopChan := make(chan bool)

	go func() {
		log.Printf("Consumer ready, PID: %d", os.Getpid())
		for d := range msgChan {
			log.Printf("Received: %s", d.Body)
			task := &TranscodeVODTask{}
			err := json.Unmarshal(d.Body, task)
			if err != nil {
				err = fmt.Errorf("Listen: failed to unmarshal json: %w", err)
				log.Printf("%+v", err)
			}
			log.Printf("encoding src to dst: %s -> %s", task.Src, task.Dst)

			// Acknowledge msg
			err = d.Ack(false)
			if err != nil {
				err = fmt.Errorf("Listen: failed to acknowledge message: %w", err)
			} else {
				log.Println("Msg acked")
			}
		}
	}()

	// Stop for program termination
	<-stopChan

	return nil
}
