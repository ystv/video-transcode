package event

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/streadway/amqp"
)

// Consumer for receiving AMPQ events
type Consumer struct {
	conn *amqp.Connection
	cdn  *s3.S3
}

// Stats represents statistics on the current encode job
type Stats struct {
	Duration   int    `json:"duration"`
	Percentage int    `json:"percentage"`
	Frame      int    `json:"frame"`
	FPS        int    `json:"fps"`
	Bitrate    string `json:"bitrate"`
	Size       string `json:"size"`
	Time       string `json:"time"`
}

// NewConsumer returns a new Consumer
func NewConsumer(conn *amqp.Connection, cdn *s3.S3) (Consumer, error) {
	c := Consumer{conn: conn, cdn: cdn}
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
	q, err := declareQueue(ch, "live")
	if err != nil {
		err = fmt.Errorf("Listen: failed to declare queue: %w", err)
		return err
	}
	msgChan, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // autoAck
		false,  // exclusive
		false,  // noLocal
		false,  // noWait
		nil,    // args
	)
	if err != nil {
		err = fmt.Errorf("Listen: failed to consume queue: %w", err)
		return err
	}

	stopChan := make(chan bool)

	go func() {
		log.Printf("VT ready, PID: %d", os.Getpid())
		for d := range msgChan {
			log.Printf("Received: %s", d.Body)
			task := &Task{}
			err := json.Unmarshal(d.Body, task)
			if err != nil {
				err = fmt.Errorf("Listen: failed to unmarshal json: %w", err)
				log.Printf("%+v", err)
			}
			err = c.TaskLive(task)
			if err != nil {
				err = fmt.Errorf("failed to transcode video: %w", err)
				log.Printf("%+v", err)
			}

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
