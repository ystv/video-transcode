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
	"github.com/ystv/video-transcode/state"
	"github.com/ystv/video-transcode/worker"
)

// Config represents VT's configuration
type Config struct {
	AMQPEndpoint       string
	CDNEndpoint        string
	CDNAccessKeyID     string
	CDNSecretAccessKey string
	StatusHost         string
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
	conf.StatusHost = os.Getenv("STATUS_HOST_BASE_URL")

	// Confirm ffmpeg installation
	cmd := exec.Command("ffmpeg", "-version")
	o, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			log.Fatalf("failed to find ffmpeg install")
		}
		log.Fatalf("failed to get ffmpeg version: %+v", err)
	}
	ver := strings.Split(string(o), " ")
	log.Println("video-transcode: v0.3.0")
	log.Printf("ffmpeg: v%s", ver[2])

	connection, err := amqp.Dial(conf.AMQPEndpoint)
	if err != nil {
		log.Fatalf("failed to connect to amqp: %+v", err)
	}
	defer connection.Close()

	var stateHandler state.ClientStateHandler = state.ClientStateHandler{}
	if err := stateHandler.Connect(conf.StatusHost); err != nil {
		log.Fatalf("failed to connect to state server: %+v", err)
	}
	defer stateHandler.Disconnect()

	consumer, err := event.NewConsumer(connection, NewCDN(), &stateHandler)
	if err != nil {
		log.Fatalf("failed to start consumer: %+v", err)
	}
	w := worker.New(consumer)
	w.Run()
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
