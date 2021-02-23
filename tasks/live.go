package task

import (
	"fmt"
	"log"
	"os/exec"
)

// TaskLive will transcode a live feed
func (ta *Tasker) TaskLive(t *Task) error {
	// TODO: ffprobe src
	cmdString := fmt.Sprintf("ffmpeg %s %s -i \"%s\" %s \"%s\" 2>&1",
		t.Args, t.SrcArgs, t.SrcURL, t.DstArgs, t.DstURL)
	// ffmpeg {glob args} {src args} -i {src url} {dst args} {dst url} 2>&1
	log.Print(cmdString)
	cmd := exec.Command("sh", "-c",
		cmdString)

	// stdout, _ := cmd.StdoutPipe()
	err := cmd.Start()
	if err != nil {
		err = fmt.Errorf("failed to start ffmpeg: %w", err)
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}
