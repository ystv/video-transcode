package main

import (
	"github.com/streadway/amqp"
	"github.com/ystv/video-transcode/event"
)

func main() {
	connection, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	if err != nil {
		panic(err)
	}
	defer connection.Close()

	consumer, err := event.NewConsumer(connection)
	if err != nil {
		panic(err)
	}
	consumer.Listen()
}
