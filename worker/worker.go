// Package worker is the main control of a worker node
package worker

import (
	"log"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
	"github.com/ystv/video-transcode/event"
)

// Worker is a control object  which listens on both
// MQ and WS
type Worker struct {
	ws *websocket.Conn
	mq *event.Consumer
}

func New(mq *event.Consumer) *Worker {
	return &Worker{mq: mq}
}

func (w *Worker) Run() {

	// w.mq.Listen()

	w.Listen(url.URL{Scheme: "ws", Host: "localhost:7071", Path: "/ws"})

	log.Printf("VT ready, PID: %d", os.Getpid())

}
