package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
)

type (
	// Tasker runs tasks in ffmpeg
	Tasker struct {
		tasks map[string]Task
		// depenendencies
		cdn *s3.S3
	}
	// Task is a generic representation of a task
	Task interface {
		GetID() string
		GetStatus() Status
		ValidateRequest() error // Generates TaskID as well as validation
		Start(ctx context.Context) error
	}
	Status struct {
		Stage      string    `json:"stage"`      // Is it downloading / transcoding / uploading
		StageStart time.Time `json:"stageStart"` // Time of when the stage started
		Stats      Stats     `json:"stats"`      // For during the transcoding stage
		Err        error     `json:"err"`        // An error inside the task
	}
)

const (
	StageStarted     string = "started"
	StageUploading   string = "uploading"
	StageTranscoding string = "transcoding"
	StageDownloading string = "downloading"
)

// New creates a task runner
func New(cdn *s3.S3) *Tasker {
	return &Tasker{tasks: make(map[string]Task), cdn: cdn}
}

// Add a task to the tasker and start
func (ta *Tasker) Add(ctx context.Context, t Task) error {
	_, exists := ta.tasks[t.GetID()]
	if exists {
		return errors.New("duplicate job id:" + t.GetID())
	}
	ta.tasks[t.GetID()] = t

	err := ta.tasks[t.GetID()].Start(ctx)
	if err != nil {
		delete(ta.tasks, t.GetID())
		return fmt.Errorf("failed to start job: %w", err)
	}
	delete(ta.tasks, t.GetID())
	return nil
}

func (ta *Tasker) GetTasks(ctx context.Context) []Task {
	tasks := []Task{}
	for task := range ta.tasks {
		tasks = append(tasks, ta.tasks[task])
	}
	return tasks
}
