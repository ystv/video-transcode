package task

import (
	"context"
	"fmt"
	"log"
	"os/exec"
)

// SimpleVideo represents a task to transcode for VOD or simple
// Essentially just basic inputs to ffmpeg
type SimpleVideo struct {
	TaskID  string `json:"taskid"`  // Task UUID
	Args    string `json:"args"`    // Global arguments
	SrcArgs string `json:"srcArgs"` // Input file options
	SrcURL  string `json:"srcURL"`  // Location of source file on CDN
	DstArgs string `json:"dstArgs"` // Output file options
	DstURL  string `json:"dstURL"`  // Destination of finished encode on CDN
}

// GetID retrives the task ID
func (t SimpleVideo) GetID() string {
	return t.TaskID
}

// Start a task
// This will only execute ffmpeg
func (t SimpleVideo) Start(ctx context.Context) error {
	// TODO: ffprobe src
	cmdString := fmt.Sprintf("ffmpeg %s %s -i \"%s\" %s \"%s\" 2>&1",
		t.Args, t.SrcArgs, t.SrcURL, t.DstArgs, t.DstURL)
	// ffmpeg {glob args} {src args} -i {src url} {dst args} {dst url} 2>&1
	log.Print(cmdString)
	cmd := exec.Command("sh", "-c",
		cmdString)

	// Put ffmpeg's log into a nice struct and log it

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}
	parseStat(cmd.StdoutPipe())

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("exec failed to wait: %w", err)
	}
	log.Println("job well done")

	return nil
}
