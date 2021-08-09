// Package worker is the main control of a worker node
package worker

import (
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ystv/video-transcode/event"
	"github.com/ystv/video-transcode/task"
)

// Config stores the settings
type Config struct {
	WorkerID     string
	APIEndpoint  string
	TasksEnabled []string
}

// Worker is a control object  which listens on both
// MQ and WS
type Worker struct {
	conf Config
	// dependencies
	task *task.Tasker
	mq   *event.Eventer
	cdn  *s3.S3
}

func New(conf Config, mq *event.Eventer, tasker *task.Tasker, cdn *s3.S3) *Worker {
	return &Worker{conf: conf, mq: mq, task: tasker, cdn: cdn}
}

func (w *Worker) Run() error {

	wg := sync.WaitGroup{}

	wg.Add(1)
	// Listen for new tasks on the message queue
	go w.Listen(&wg)
	// if err != nil {
	// 	log.Printf("failed to listen: %+v", err)
	// }

	wg.Add(1)
	// Send back status to the message queue
	go w.PubStatus(&wg)
	// if err != nil {
	// 	log.Printf("failed to publish status: %+v", err)
	// }

	log.Printf("VT ready, worker ID: %s, PID: %d", w.conf.WorkerID, os.Getpid())

	wg.Wait()

	return nil
}
