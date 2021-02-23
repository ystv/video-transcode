package task

import (
	"io"
	"log"

	"github.com/aws/aws-sdk-go/service/s3"
)

type (
	// Tasker runs tasks in ffmpeg
	Tasker struct {
		cdn *s3.S3
		log *log.Logger
	}
	// Task represents a task to transcode for VOD or raw
	// Essentially just basic inputs to ffmpeg
	Task struct {
		ID      string `json:"id"`      // Task UUID
		Args    string `json:"args"`    // Global arguments
		SrcArgs string `json:"srcArgs"` // Input file options
		SrcURL  string `json:"srcURL"`  // Location of source file on CDN
		DstArgs string `json:"dstArgs"` // Output file options
		DstURL  string `json:"dstURL"`  // Destination of finished encode on CDN
	}
)

// New creates a task runner
func New(cdn *s3.S3, logHandle io.Writer) *Tasker {
	l := log.New(logHandle, "node: a task: 100: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	return &Tasker{cdn: cdn, log: l}
}
