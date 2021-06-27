package manager

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/ystv/video-transcode/state"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Upgrade will convert a request into a websocket
func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade: %w", err)
	}
	return ws, nil
}

// Reader reads the websocket connection
func (m *Manager) Reader(conn *websocket.Conn) {
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		var updateMessage state.StatusUpdate
		if err := json.Unmarshal(p, &updateMessage); err != nil {
			log.Println(err)
			return
		}

		switch updateMessage.Header {
		case "WORKER":
			var wStatus state.WorkerStatusUpdate
			byt, _ := json.Marshal(updateMessage.Body)
			json.Unmarshal(byt, &wStatus)

			switch wStatus.State {
			case "START":
				m.state.Workers[wStatus.WorkerID] = &state.WorkerStatus{}
			case "END":
				delete(m.state.Workers, wStatus.WorkerID)
			case "ADD JOB":
				m.state.Workers[wStatus.WorkerID].StartJob()
			case "END JOB":
				m.state.Workers[wStatus.WorkerID].EndJob()
			}
		case "JOB":
			var jStatus state.FullStatusIndicator
			byt, _ := json.Marshal(updateMessage.Body)
			json.Unmarshal(byt, &jStatus)

			m.state.Jobs[jStatus.GetUUID()] = jStatus
		}

		log.Println(string(p))
	}
}

// Writer writes into the websocket
func (m *Manager) Writer(conn *websocket.Conn) {
	for {
		fmt.Println("Sending")
		messageType, r, err := conn.NextReader()
		if err != nil {
			fmt.Println(err)
			return
		}
		w, err := conn.NextWriter(messageType)
		if err != nil {
			fmt.Println(err)
			return
		}
		if _, err := io.Copy(w, r); err != nil {
			fmt.Println(err)
			return
		}
		if err := w.Close(); err != nil {
			fmt.Println(err)
			return
		}
	}
}
