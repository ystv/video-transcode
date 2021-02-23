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
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"github.com/ystv/video-transcode/event"
)

// Config represents VT's configuration
type Config struct {
	AMQPEndpoint       string
	CDNEndpoint        string
	CDNAccessKeyID     string
	CDNSecretAccessKey string
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

	connection, err := amqp.Dial(conf.AMQPEndpoint)
	if err != nil {
		err = fmt.Errorf("failed to connect to amqp: %w", err)
		log.Fatalf("%+v", err)
	}
	defer connection.Close()

	m, err := NewManager()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	consumer, err := event.NewConsumer(connection, NewCDN(), m)
	if err != nil {
		panic(err)
	}
	consumer.Listen()
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

// NewManager creates a ws connection with the manager node
// allows metrics and remote context
func NewManager() (*websocket.Conn, error) {
	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:7071/ws", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ws: %w", err)
	}
	defer c.Close()
	return c, nil
}
