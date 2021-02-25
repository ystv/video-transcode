// Package worker is the main control of a worker node
package worker

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
	"github.com/ystv/video-transcode/event"
	"github.com/ystv/video-transcode/task"
)

// Worker is a control object  which listens on both
// MQ and WS
type Worker struct {
	state map[string]*task.Task
	// dependencies
	ws *websocket.Conn
	mq *event.Consumer
}

func New(mq *event.Consumer) *Worker {
	return &Worker{mq: mq}
}

func (w *Worker) Run() error {
	err := w.mq.Listen()
	if err != nil {
		return fmt.Errorf("mq failed: %w", err)
	}

	w.Listen(url.URL{Scheme: "ws", Host: "localhost:7071", Path: "/ws"})

	log.Printf("VT ready, PID: %d", os.Getpid())

	return nil
}
