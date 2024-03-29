package task

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const TypeVOD string = "video/vod"

var _ Task = &VOD{}

// VOD task produces a video for the on demand platform
type VOD struct {
	TaskID  string // Task UUID
	SrcURL  string `json:"srcURL"`  // Location of source file on CDN
	DstArgs string `json:"dstArgs"` // Output file options
	DstURL  string `json:"dstURL"`  // Destination of finished encode on CDN

	APIEndpoint string

	status Status
	stats  *Stats

	// dependencies
	cdn *s3.S3
}

// NewVOD initialises a VOD task object so we can
// add the tasks dependencies
func NewVOD(cdn *s3.S3, apiEndpoint string) VOD {
	return VOD{
		status:      Status{},
		stats:       &Stats{},
		cdn:         cdn,
		APIEndpoint: apiEndpoint,
	}
}

// GetID returns a task ID
func (t *VOD) GetID() string {
	return t.TaskID
}

func (t *VOD) GetStatus() Status {
	return t.status
}

// CheckRequets returns an error describing if the user's request is not
// formed properly and will stop the job continuing
func (t *VOD) ValidateRequest() error {
	if t.SrcURL == "" {
		return fmt.Errorf("missing srcURL")
	}
	if t.DstURL == "" {
		return fmt.Errorf("missing dstURL")
	}
	return nil
}

// Start makes a video for VOD
//
// General outline
// Creates a temp file
// Creates a new S3 session
// Download video object from S3 and put it in temp file
// Execute ffmpeg arguements on downloaded file
// Upload result file
func (t *VOD) Start(ctx context.Context) error {
	t.status.Stage = StageStarted
	t.status.StageStart = time.Now()

	// Change slashes with dashes making it easier to handle in the FS
	srcPath := strings.Split(t.SrcURL, "/")
	dstPath := strings.Split(t.DstURL, "/")
	dstFilename := strings.Join(dstPath[1:], "-")

	url, err := t.presignFileURL(&srcPath[0], aws.String(strings.Join(srcPath[1:], "/")))
	if err != nil {
		return fmt.Errorf("failed to sign source download: %w", err)
	}

	// Video encoding
	log.Printf("encoding video: %s", t.GetID())
	startEnc := time.Now()
	t.status.Stage = StageTranscoding
	t.status.StageStart = startEnc

	// TODO More Status Updates Below This Point

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

	log.Printf("%+v", t)
	log.Println(cmdString)

	// begin encoding
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

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
		return fmt.Errorf("exec failed to wait: %+v: %s", err, curLine)
	}

	log.Printf("finished encoding - completed in %s", time.Since(startEnc))
	startUp := time.Now()
	t.status.Stage = StageUploading
	t.status.StageStart = startUp

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

	c := http.Client{}

	res, err := c.Post(os.Getenv("VT_WAPI_ENDPOINT")+"/v1/internal/encoder/transcode_finished/"+t.TaskID, "", nil)
	if err != nil {
		return "", fmt.Errorf("failed to post to vt: %w", err)
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to send complete status to web-api")
	}
	log.Println("uploaded video!")

	return upload.Location, nil
}
