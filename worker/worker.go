// Package worker is the main control of a worker node
package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

// Worker is a control object  which listens on both
// MQ and WS
type Worker struct {
	ws *websocket.Conn
}

// Listen to a websocket
func (w *Worker) Listen(u url.URL) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	var err error
	w.ws, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer w.ws.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := w.ws.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := w.ws.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := w.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

// Announce sends a message to the manager
func (w *Worker) Announce(ffmpegVersion string) error {
	announce := struct {
		FFmpegVersion string
		VTVersion     string
		Bandwidth     int // could run quick speed test
	}{
		FFmpegVersion: ffmpegVersion,
	}
	msg, err := json.Marshal(announce)
	if err != nil {
		return fmt.Errorf("failed to marhsal announce: %w", err)
	}

	w.ws.WriteMessage(websocket.TextMessage, []byte(msg))
	return nil
}
