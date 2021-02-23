package manager

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	task "github.com/ystv/video-transcode/tasks"
	"github.com/ystv/video-transcode/ws"
)

// Router encapsulates the managers HTTP endpoints
func (m *Manager) Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", m.indexHandle)
	r.HandleFunc("/ok", m.healthHandle)
	r.HandleFunc("/task/vod", m.newVODHandle)
	r.HandleFunc("/task/raw", m.newLiveHandle)
	r.HandleFunc("/ws", m.newWS)
	return r
}

// indexHandle just shows it's alive, could have metrics?
func (m *Manager) indexHandle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("vt manager (v0.3.0)"))
}

// healthHandle for other services to check it's healthy
func (m *Manager) healthHandle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// newVODHandle will download file from CDN to local
// transcode, upload and cleanup
func (m *Manager) newVODHandle(w http.ResponseWriter, r *http.Request) {
	t := task.Task{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = m.mq.Push(t, "vod")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("encoding"))
}

// newLiveHandle will stream file to ffmpeg, optional
// transcode and send to new endpoint
func (m *Manager) newLiveHandle(w http.ResponseWriter, r *http.Request) {
	t := task.Task{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = m.mq.Push(t, "live")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("encoding"))
}

// newWS handles upgrading a worker node's connection to a ws.
// Used for metrics and context
func (m *Manager) newWS(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+V\n", err)
		return
	}
	go ws.Writer(conn)
	ws.Reader(conn)
}
