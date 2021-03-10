package manager

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ystv/video-transcode/task"
)

// Router encapsulates the managers HTTP endpoints
func (m *Manager) Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", m.indexHandle)
	r.HandleFunc("/ok", m.healthHandle)
	r.HandleFunc("/task/image/simple", m.newImageSimple)
	r.HandleFunc("/task/video/simple", m.newVideoSimpleHandle)
	r.HandleFunc("/task/video/vod", m.newVideoOnDemandHandle)
	r.HandleFunc("/task/video/probe", m.indexHandle)
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

func (m *Manager) newImageSimple(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("simple image"))
}

// newVideoOnDemandHandle will download file from CDN to local
// transcode, upload and cleanup
func (m *Manager) newVideoOnDemandHandle(w http.ResponseWriter, r *http.Request) {
	t := task.VOD{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = m.mq.Push(t, "video/vod")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("encoding"))
}

// newLiveHandle will stream file to ffmpeg, optional
// transcode and send to new endpoint
func (m *Manager) newVideoSimpleHandle(w http.ResponseWriter, r *http.Request) {
	t := task.SimpleVideo{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = m.mq.Push(t, "video/simple")
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
	conn, err := Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+V\n", err)
		return
	}
	go Writer(conn)
	Reader(conn)
}
