package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/ystv/video-transcode/event"
	"github.com/ystv/video-transcode/task"
)

// Listen will listen for all new Queue publications
// and print them to the console.
func (w *Worker) Listen(wg *sync.WaitGroup) error {

	defer wg.Done()

	ch, err := w.mq.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	defer ch.Close()

	msgChan, err := event.ListenToQueues(ch, w.conf.TasksEnabled)
	if err != nil {
		return fmt.Errorf("failed to start listening channels")
	}

	// Going through all deliveries
	for d := range msgChan {
		switch d.RoutingKey {
		case task.TypeVOD:
			log.Println("video/vod job received!")
			t := task.NewVOD(w.cdn, w.conf.APIEndpoint)
			err := json.Unmarshal(d.Body, &t)
			if err != nil {
				err = fmt.Errorf("failed to unmarshal json: %w", err)
				log.Printf("%+v", err)
			}
			err = w.task.Add(context.Background(), &t)
			if err != nil {
				err = fmt.Errorf("failed to add job: %w", err)
				log.Printf("%+v", err)
			}

		case task.TypeSimpleVideo:
			log.Println("video/simple job received!")
			t := task.SimpleVideo{}
			err := json.Unmarshal(d.Body, &t)
			if err != nil {
				err = fmt.Errorf("failed to unmarshal json: %w", err)
				log.Printf("%+v", err)
			}
			err = w.task.Add(context.Background(), &t)
			if err != nil {
				err = fmt.Errorf("failed to add job: %w", err)
				log.Printf("%+v", err)
			}
			log.Println("job added to task manager!")
		}
		// Acknowledge msg
		err := d.Ack(false)
		if err != nil {
			err = fmt.Errorf("failed to acknowledge message: %w", err)
		}
		log.Println("job well done lads")
	}
	log.Println("that'll do")
	return nil
}
