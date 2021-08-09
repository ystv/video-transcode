package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"github.com/ystv/video-transcode/event"
	"github.com/ystv/video-transcode/task"
	"github.com/ystv/video-transcode/worker"
)

// Config represents VT's configuration
type Config struct {
	AMQPEndpoint       string
	CDNEndpoint        string
	CDNAccessKeyID     string
	CDNSecretAccessKey string
	APIEndpoint        string
}

var conf Config

func main() {
	// Initialising config
	godotenv.Load(".env.local")
	godotenv.Load(".env")
	conf.AMQPEndpoint = os.Getenv("VT_AMQP_ENDPOINT")
	conf.CDNEndpoint = os.Getenv("VT_CDN_ENDPOINT")
	conf.CDNAccessKeyID = os.Getenv("VT_CDN_ACCESSKEYID")
	conf.CDNSecretAccessKey = os.Getenv("VT_CDN_SECRETACCESSKEY")
	conf.APIEndpoint = os.Getenv("VT_WAPI_ENDPOINT")

	// Confirm ffmpeg installation
	output, err := exec.Command("ffmpeg", "-version").Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			log.Fatalf("failed to find ffmpeg install")
		}
		log.Fatalf("failed to get ffmpeg version: %+v", err)
	}
	ver := strings.Split(string(output), " ")
	log.Println("video-transcode: v0.3.0")
	log.Printf("ffmpeg: v%s", ver[2])

	conn, err := amqp.Dial(conf.AMQPEndpoint)
	if err != nil {
		log.Fatalf("failed to connect to amqp: %+v", err)
	}
	defer conn.Close()

	cdn := NewCDN()
	eventer, err := event.NewEventer(conn)
	if err != nil {
		log.Fatalf("failed to create new eventer: %+v", err)
	}

	wConf := worker.Config{
		WorkerID:     "test-worker",
		APIEndpoint:  conf.APIEndpoint,
		TasksEnabled: []string{"video/simple", "video/vod"}}

	w := worker.New(wConf, eventer, task.New(cdn), cdn)
	err = w.Run()
	if err != nil {
		log.Fatalf("failed to run worker: %+v", err)
	}
}

// NewCDN creates a connection to s3
func NewCDN() *s3.S3 {
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			conf.CDNAccessKeyID,
			conf.CDNSecretAccessKey, ""),
		Endpoint:         aws.String(conf.CDNEndpoint),
		Region:           aws.String("ystv-wales-1"),
		S3ForcePathStyle: aws.Bool(true),
	}
	newSession := session.New(s3Config)
	return s3.New(newSession)
}
