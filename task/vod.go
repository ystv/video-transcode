package task

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Stats represents statistics on the current encode job
type Stats struct {
	Duration   int    `json:"duration"`
	Percentage int    `json:"percentage"`
	Frame      int    `json:"frame"`
	FPS        int    `json:"fps"`
	Bitrate    string `json:"bitrate"`
	Size       string `json:"size"`
	Time       string `json:"time"`
}

// TaskVOD makes a video for VOD
//
// General outline
// Creates a temp file
// Creates a new S3 session
// Download video object from S3 and put it in temp file
// Execute ffmpeg arguements on downloaded file
// Upload result file
func (ta *Tasker) TaskVOD(t *Task) error {
	// Change slashes with dashes making it easier to handle in the FS
	srcPath := strings.Split(t.SrcURL, "/")
	dstPath := strings.Split(t.DstURL, "/")
	dstFilename := strings.Join(dstPath[1:], "-")

	url, err := ta.presignFileURL(&srcPath[0], aws.String(strings.Join(srcPath[1:], "/")))
	if err != nil {
		return fmt.Errorf("failed to sign source download: %w", err)
	}

	// Video encoding
	log.Printf("encoding video")
	startEnc := time.Now()

	// We're not using the -progress flag since it doesn't give us the duration
	// of the video which is important to determine the ETA. so we'll just parsing
	// the normal stdout.
	cmdString := fmt.Sprintf("%s \"%s\" %s \"%s\" %s",
		"ffmpeg -y -i", url, t.DstArgs, dstFilename, "2>&1")

	cmd := exec.Command("sh", "-c",
		cmdString)

	// Put ffmpeg's log into a nice struct and log it
	go parseStat(cmd.StdoutPipe())

	// begin encoding
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("exec failed: %+v", err)
	}

	log.Printf("finished encoding - completed in %s", time.Since(startEnc))
	startUp := time.Now()

	// Uploading encoded file
	_, err = ta.uploadFile(dstFilename, dstPath)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	log.Printf("finished uploading - completed in %s", time.Since(startUp))

	return nil
}

func (ta *Tasker) presignFileURL(bucket, key *string) (string, error) {
	req, _ := ta.cdn.GetObjectRequest(&s3.GetObjectInput{
		Bucket: bucket,
		Key:    key,
	})
	return req.Presign(6 * time.Hour) // TODO: Look into time
}

func (ta *Tasker) uploadFile(src string, dst []string) (string, error) {
	file, err := os.Open(src)
	if err != nil {
		err = fmt.Errorf("failed to open encoded file: %w", err)
		return "", err
	}
	defer file.Close()
	sess, err := session.NewSession(&ta.cdn.Config)
	if err != nil {
		return "", fmt.Errorf("failed to create new cdn session: %w", err)
	}
	uploader := s3manager.NewUploader(sess)
	upload, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(dst[0]),
		Key:    aws.String(strings.Join(dst[1:], "/")),
		Body:   file,
	})
	if err != nil {
		err = fmt.Errorf("failed to upload encoded file: %w", err)
		return "", err
	}
	file.Close()

	// Deleting local encoded file
	err = os.Remove(src)
	if err != nil {
		err = fmt.Errorf("failed to delete source file: %w", err)
		return "", err
	}
	return upload.Location, nil
}

func parseStat(stdout io.ReadCloser, err error) error {
	bytes := make([]byte, 100)
	stats := &Stats{}
	allRes := ""
	for {
		_, err := stdout.Read(bytes)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("failed to read stdout: %w", err)
		}
		allRes += string(bytes)
		ok := getStats(allRes, stats)
		if ok {
			allRes = ""
			log.Printf("%+v", stats)
		}
	}
	return nil
}

func getStats(res string, s *Stats) bool {

	durIdx := strings.Index(res, "Duration")
	// Checking if we've got a "Duration",
	// we need this so we can determine the ETA
	if durIdx >= 0 {

		dur := res[durIdx+10:]
		if len(dur) > 8 {
			dur = dur[0:8]

			s.Duration = durToSec(dur)
			return true
		}
	}
	// FFmpeg should give us a duration on startup,
	// so we kill here in the event that didn't happen.
	if s.Duration == 0 {
		return false
	}

	frameIdx := strings.LastIndex(res, "frame=")
	fpsIdx := strings.LastIndex(res, "fps=")
	bitrateIdx := strings.LastIndex(res, "bitrate=")
	sizeIdx := strings.LastIndex(res, "size=")
	timeIdx := strings.Index(res, "time=")

	if timeIdx >= 0 {
		// From this point on it should be outputting normal encode stdout,
		// which we'll want to parse.

		frame := strings.Fields(res[frameIdx+6:])
		fps := strings.Fields(res[fpsIdx+4:])
		bitrate := strings.Fields(res[bitrateIdx+8:])
		size := strings.Fields(res[sizeIdx+5:])
		time := res[timeIdx+5:]

		if len(time) > 8 {
			time = time[0:8]
			sec := durToSec(time)
			per := (sec * 100) / s.Duration
			if s.Percentage != per {
				s.Percentage = per
				// Just doing to reuse this int variable for each item
				integer, _ := strconv.Atoi(frame[0])
				s.Frame = integer
				integer, _ = strconv.Atoi(fps[0])
				s.FPS = integer
				s.Bitrate = bitrate[0]
				s.Size = size[0]
				s.Time = time
			}
			return true
		}
	}
	return false
}

func durToSec(dur string) (sec int) {
	// So we're kind of limiting our videos to 24h which isn't ideal
	// shouldn't crash the application hopefully XD
	durAry := strings.Split(dur, ":")
	if len(durAry) != 3 {
		return
	}
	hr, _ := strconv.Atoi(durAry[0])
	sec = hr * (60 * 60)
	min, _ := strconv.Atoi(durAry[1])
	sec += min * (60)
	second, _ := strconv.Atoi(durAry[2])
	sec += second
	return
}
