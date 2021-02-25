package task

import (
	"context"
	"errors"
	"io"
	"log"

	"github.com/aws/aws-sdk-go/service/s3"
)

type (
	// Tasker runs tasks in ffmpeg
	Tasker struct {
		active map[string]Task
		// depenendencies
		cdn *s3.S3
		log *log.Logger
	}
	// Task is a generic representation of a task
	Task interface {
		GetID() string
		Start(ctx context.Context) error
	}
	// Stats represents statistics on the current encode job
	Stats struct {
		Duration   int    `json:"duration"`
		Percentage int    `json:"percentage"`
		Frame      int    `json:"frame"`
		FPS        int    `json:"fps"`
		Bitrate    string `json:"bitrate"`
		Size       string `json:"size"`
		Time       string `json:"time"`
	}
)

// New creates a task runner
func New(cdn *s3.S3, logHandle io.Writer) *Tasker {
	l := log.New(logHandle, "node: a task: 100: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	return &Tasker{cdn: cdn, log: l}
}

// Add a task to the tasker and start
func (ta *Tasker) Add(ctx context.Context, t Task) error {
	_, exists := ta.active[t.GetID()]
	if exists {
		return errors.New("duplicate job id:" + t.GetID())
	}
	ta.active[t.GetID()] = t

	ta.active[t.GetID()].Start(ctx)
	return nil
}
