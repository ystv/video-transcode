package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/streadway/amqp"
	"github.com/ystv/video-transcode/state"
	"github.com/ystv/video-transcode/task"
)

// Consumer for receiving AMPQ events
type Consumer struct {
	conn         *amqp.Connection
	cdn          *s3.S3
	task         *task.Tasker
	stateHandler *state.ClientStateHandler
}

// NewConsumer returns a new Consumer
func NewConsumer(conn *amqp.Connection, cdn *s3.S3, stateHandler *state.ClientStateHandler) (*Consumer, error) {
	c := &Consumer{conn: conn, cdn: cdn, task: task.New(cdn), stateHandler: stateHandler}
	ch, err := c.conn.Channel()
	if err != nil {
		err = fmt.Errorf("NewConsumer: failed to get channel: %w", err)
		return nil, err
	}
	err = declareExchange(ch)
	if err != nil {
		err = fmt.Errorf("NewConsumer: failed to declare exchange: %w, err", err)
		return nil, err
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

	msgChan, err := newListener(ch, []string{"video/vod", "video/simple"})
	if err != nil {
		return fmt.Errorf("failed to start listening channels")
	}

	stopChan := make(chan bool)

	go func() {
		for d := range msgChan {
			switch d.RoutingKey {
			case queueVideoVOD:
				log.Println("video/vod job")
				t := task.NewVOD(c.cdn)
				err := json.Unmarshal(d.Body, &t)
				if err != nil {
					err = fmt.Errorf("Listen: failed to unmarshal json: %w", err)
					log.Printf("%+v", err)
				}
				err = c.task.Add(context.Background(), &t, c.stateHandler)
				if err != nil {
					err = fmt.Errorf("failed to add job: %w", err)
					log.Printf("%+v", err)
				}

			case queueVideoSimple:
				log.Println("video/simple job")
				t := task.SimpleVideo{}
				err := json.Unmarshal(d.Body, &t)
				if err != nil {
					err = fmt.Errorf("Listen: failed to unmarshal json: %w", err)
					log.Printf("%+v", err)
				}
				err = c.task.Add(context.Background(), &t, c.stateHandler)
				if err != nil {
					err = fmt.Errorf("failed to add job: %w", err)
					log.Printf("%+v", err)
				}
			case queueImageSimple:
				log.Println("image/simple job")
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
