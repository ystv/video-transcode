package task

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os/exec"

	"github.com/google/uuid"
)

// SimpleVideo represents a task to transcode for VOD or simple
// Essentially just basic inputs to ffmpeg
type SimpleVideo struct {
	TaskID  string `json:"taskid"`  // Task UUID
	Args    string `json:"args"`    // Global arguments
	SrcArgs string `json:"srcArgs"` // Input file options
	SrcURL  string `json:"srcURL"`  // Location of source file on CDN
	DstArgs string `json:"dstArgs"` // Output file options
	DstURL  string `json:"dstURL"`  // Destination of finished encode on

	stats *Stats
}

// GetID retrives the task ID
func (t *SimpleVideo) GetID() string {
	return t.TaskID
}

// CheckRequets returns an error describing if the user's request is not
// formed properly and will stop the job continuing
func (t *SimpleVideo) ValidateRequest() error {
	if t.SrcURL == "" {
		return fmt.Errorf("missing srcURL")
	}
	if t.DstURL == "" {
		return fmt.Errorf("missing dstURL")
	}

	// Generating Task ID
	t.TaskID = uuid.NewString()
	return nil
}

// Start a task
// This will only execute ffmpeg
func (t *SimpleVideo) Start(ctx context.Context) error {
	// TODO: ffprobe src
	cmdString := fmt.Sprintf("ffmpeg %s %s -i \"%s\" %s \"%s\" 2>&1",
		t.Args, t.SrcArgs, t.SrcURL, t.DstArgs, t.DstURL)
	// ffmpeg {glob args} {src args} -i {src url} {dst args} {dst url} 2>&1
	cmd := exec.Command("sh", "-c",
		cmdString)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("pipe failed: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	t.stats = &Stats{}

	scanner := bufio.NewScanner(stdout)
	curLine := ""
	buf := ""

	for scanner.Scan() {
		curLine = scanner.Text()
		buf += curLine
		ok := getStats(t.stats, buf)
		if ok {
			buf = ""
			log.Printf("%+v", t.stats)
		}
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("exec failed to wait: %w: %s", err, curLine)
	}
	return nil
}
