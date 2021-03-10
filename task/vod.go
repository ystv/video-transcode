package task

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var _ Task = &VOD{}

// VOD task produces a video for the on demand platform
type VOD struct {
	TaskID  string `json:"taskID"`  // Task UUID
	Args    string `json:"args"`    // Global arguments
	SrcArgs string `json:"srcArgs"` // Input file options
	SrcURL  string `json:"srcURL"`  // Location of source file on CDN
	DstArgs string `json:"dstArgs"` // Output file options
	DstURL  string `json:"dstURL"`  // Destination of finished encode on CDN

	stats *Stats

	// dependencies
	cdn *s3.S3
}

// NewVOD initialises a VOD task object so we can
// add the tasks dependencies
func NewVOD(cdn *s3.S3) VOD {
	return VOD{cdn: cdn}
}

// GetID returns a task ID
func (t VOD) GetID() string {
	return t.TaskID
}

// Start makes a video for VOD
//
// General outline
// Creates a temp file
// Creates a new S3 session
// Download video object from S3 and put it in temp file
// Execute ffmpeg arguements on downloaded file
// Upload result file
func (t VOD) Start(ctx context.Context) error {
	// Change slashes with dashes making it easier to handle in the FS
	srcPath := strings.Split(t.SrcURL, "/")
	dstPath := strings.Split(t.DstURL, "/")
	dstFilename := strings.Join(dstPath[1:], "-")

	url, err := t.presignFileURL(&srcPath[0], aws.String(strings.Join(srcPath[1:], "/")))
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

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("pipe failed: %w", err)
	}

	// begin encoding
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
		return fmt.Errorf("exec failed: %w", err)
	}

	log.Printf("finished encoding - completed in %s", time.Since(startEnc))
	startUp := time.Now()

	// Uploading encoded file
	_, err = t.uploadFile(dstFilename, dstPath)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	log.Printf("finished uploading - completed in %s", time.Since(startUp))

	return nil
}

func (t *VOD) presignFileURL(bucket, key *string) (string, error) {
	req, _ := t.cdn.GetObjectRequest(&s3.GetObjectInput{
		Bucket: bucket,
		Key:    key,
	})
	return req.Presign(6 * time.Hour) // TODO: Look into time
}

func (t *VOD) uploadFile(src string, dst []string) (string, error) {
	file, err := os.Open(src)
	if err != nil {
		err = fmt.Errorf("failed to open encoded file: %w", err)
		return "", err
	}
	defer file.Close()
	sess, err := session.NewSession(&t.cdn.Config)
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
