package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/streadway/amqp"
	"github.com/ystv/video-transcode/event"
)

func main() {
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
	log.Println("video-transcode: v0.1.0")
	log.Printf("ffmpeg: v%s", ver[2])

	connection, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	if err != nil {
		err = fmt.Errorf("failed to connect to amqp: %w", err)
		log.Fatalf("%+v", err)
	}
	defer connection.Close()

	consumer, err := event.NewConsumer(connection, NewCDN())
	if err != nil {
		panic(err)
	}
	consumer.Listen()
}

// NewCDN creates a connection to s3
func NewCDN() *s3.S3 {
	endpoint := os.Getenv("CDN_ENDPOINT")
	accessKeyID := os.Getenv("CDN_ACCESSKEYID")
	secretAccessKey := os.Getenv("CDN_SECRETACCESSKEY")

	// Configure to use CDN Server

	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String("ystv-wales-1"),
		S3ForcePathStyle: aws.Bool(true),
	}
	newSession := session.New(s3Config)
	return s3.New(newSession)
}
