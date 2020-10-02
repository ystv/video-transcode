package main

import (
	"log"

	"github.com/streadway/amqp"
	"github.com/ystv/video-transcode/event"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	if err != nil {
		panic(err)
	}

	emitter, err := event.NewProducer(conn)
	if err != nil {
		panic(err)
	}

	task := event.TranscodeVODTask{
		Src:        "videos/American Football Match.mp4",
		Dst:        "videos/2020_AFM_sum.mp4",
		EncodeName: "FHD 8Mbps",
		EncodeArgs: "-vf scale=1920:1080 -c:v libx264 -crf 20 -preset slow -c:a copy -threads 0",
	}

	for i := 1; i < 10; i++ {
		err = emitter.Push(task)
		if err != nil {
			log.Printf("%+v", err)
		} else {
			log.Print("task sent")
		}
	}
}
