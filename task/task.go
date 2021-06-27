package task

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ystv/video-transcode/state"
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
		ValidateRequest() error // Generates TaskID as well as validation
		Start(ctx context.Context, sh *state.ClientStateHandler) error
	}
)

// New creates a task runner
func New(cdn *s3.S3) *Tasker {
	return &Tasker{tasks: make(map[string]Task), cdn: cdn}
}

// Add a task to the tasker and start
func (ta *Tasker) Add(ctx context.Context, t Task, sh *state.ClientStateHandler) error {
	_, exists := ta.tasks[t.GetID()]
	if exists {
		return errors.New("duplicate job id:" + t.GetID())
	}
	ta.tasks[t.GetID()] = t

	sh.SendWorkerUpdate("ADD JOB")
	defer sh.SendWorkerUpdate("END JOB")
	err := ta.tasks[t.GetID()].Start(ctx, sh)
	if err != nil {
		delete(ta.tasks, t.GetID())
		sh.SendJobUpdate(state.FullStatusIndicator{
			JobID:       t.GetID(),
			FailureMode: "FAILED",
			Summary:     "Failed During Job",
			Detail:      err.Error(),
		})
		return fmt.Errorf("failed to start job: %w", err)
	}
	delete(ta.tasks, t.GetID())
	return nil
}
