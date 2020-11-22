package event

import (
	"fmt"
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

// TaskVOD makes a video for VOD
//
// General outline
// Creates a temp file
// Creates a new S3 session
// Download video object from S3 and put it in temp file
// Execute ffmpeg arguements on downloaded file
// Upload result file
func (c *Consumer) TaskVOD(payload *Task) error {
	// TODO make ffmpeg use a signed url instead of downloading the file
	// Download src
	startDl := time.Now()
	srcPath := strings.Split(payload.SrcURL, "/")
	srcFilename := strings.Join(srcPath[1:], "-")
	dstPath := strings.Split(payload.DstURL, "/")
	dstFilename := strings.Join(dstPath[1:], "-")

	file, err := os.Create(srcFilename)
	if err != nil {
		err = fmt.Errorf("failed to create temp source file: %w", err)
		return err
	}
	defer file.Close()
	sess, err := session.NewSession(&c.cdn.Config)
	if err != nil {
		err = fmt.Errorf("failed to create new cdn session: %w", err)
		return err
	}
	downloader := s3manager.NewDownloader(sess)
	numBytes, err := downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(srcPath[0]),
		Key:    aws.String(strings.Join(srcPath[1:], "/")),
	})
	if err != nil {
		err = fmt.Errorf("failed to download file: %w", err)
		return err
	}
	file.Close()
	log.Printf("downloaded %d bytes in %s", numBytes, time.Since(startDl))

	// Video encoding
	log.Printf("encode video: %s/%s", payload.SrcURL, payload.DstURL)

	cmdString := fmt.Sprintf("%s \"%s\" %s \"%s\" %s",
		"ffmpeg -y -i", srcFilename, payload.DstArgs, dstFilename, "2>&1")

	cmd := exec.Command("sh", "-c",
		cmdString)

	stdout, _ := cmd.StdoutPipe()
	err = cmd.Start()
	if err != nil {
		err = fmt.Errorf("failed to start ffmpeg: %w", err)
		return err
	}

	// We're not using the -progress flag since it doesn't give us the duration
	// of the video which is important to determine the ETA. so we'll just parsing
	// the normal stdout.

	bytes := make([]byte, 100)
	stats := &Stats{}
	allRes := ""
	startEnc := time.Now()
	for {
		_, err := stdout.Read(bytes)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			err = fmt.Errorf("failed to read stdout: %w", err)
			return err
		}
		allRes += string(bytes)
		ok := getStats(allRes, stats)
		if ok {
			allRes = ""
			log.Printf("%+v", stats)
		}
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	log.Printf("finished encoding %s/%s - completed in %s", payload.SrcURL, payload.DstURL, time.Since(startEnc))

	// Deleting local source file
	err = os.Remove(srcFilename)
	if err != nil {
		err = fmt.Errorf("failed to delete source file: %w", err)
		return err
	}

	startUp := time.Now()

	// Uploading encoded file
	file, err = os.Open(dstFilename)
	if err != nil {
		err = fmt.Errorf("failed to open encoded file: %w", err)
		return err
	}
	defer file.Close()
	uploader := s3manager.NewUploader(sess)
	upload, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(dstPath[0]),
		Key:    aws.String(strings.Join(dstPath[1:], "/")),
		Body:   file,
	})
	if err != nil {
		err = fmt.Errorf("failed to upload encoded file: %w", err)
		return err
	}
	file.Close()

	log.Printf("finished uploading %s/%s to %s - completed in %s", payload.SrcURL, payload.DstURL, upload.Location, time.Since(startUp))

	// Deleting local encoded file
	err = os.Remove(dstFilename)
	if err != nil {
		err = fmt.Errorf("failed to delete source file: %w", err)
		return err
	}

	log.Printf("Finished %s/%s - completed in %s", payload.SrcURL, payload.DstURL, time.Since(startDl))

	return nil
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
