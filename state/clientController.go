package state

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ClientStateHandler struct {
	WorkerID string
	ws       *websocket.Conn
}

func (h *ClientStateHandler) Connect(url string) error {
	var err error
	h.ws, _, err = websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}

	h.WorkerID = uuid.NewString()

	return h.SendWorkerUpdate("START")

}

func (h *ClientStateHandler) Disconnect() {
	h.SendWorkerUpdate("END")
	h.ws.Close()
}

func (h *ClientStateHandler) SendWorkerUpdate(status string) error {
	return h.ws.WriteJSON(
		StatusUpdate{
			Header: "WORKER",
			Body: WorkerStatusUpdate{
				WorkerID: h.WorkerID,
				State:    status,
			}})
}

func (h *ClientStateHandler) SendJobUpdate(status FullStatusIndicator) error {
	return h.ws.WriteJSON(
		StatusUpdate{
			Header: "JOB",
			Body:   status,
		})
}
