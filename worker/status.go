package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ystv/video-transcode/task"
)

type Status struct {
	WorkerID     string
	TasksEnabled []string
	CurrentTasks []task.Status
}

func (w *Worker) PubStatus(wg *sync.WaitGroup) error {
	status := Status{
		WorkerID:     w.conf.WorkerID,
		TasksEnabled: w.conf.TasksEnabled,
	}

	defer wg.Done()

	var err error
	var reqJSON []byte
	t := time.Tick(2 * time.Second)
	for {
		select {
		case <-t:
			for _, task := range w.task.GetTasks(context.Background()) {
				status.CurrentTasks = append(status.CurrentTasks, task.GetStatus())
				log.Printf("%+v", err)
			}
			reqJSON, err = json.Marshal(status)
			if err != nil {
				return fmt.Errorf("failed to marshal: %w", err)
			}
			err = w.mq.SendStatus("encode-status", reqJSON)
			if err != nil {
				return fmt.Errorf("SendStatus failed: %w", err)
			}
		}
	}
}
