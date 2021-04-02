package task

import (
	"context"
	"errors"
	"fmt"

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
		ValidateRequest() error
		Start(ctx context.Context) error
	}
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
