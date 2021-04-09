package manager

import (
	"crypto/subtle"
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
	r.HandleFunc("/status/job/{uuid}", m.basicAuth(m.jobStateHandle))
	r.HandleFunc("/status/worker", m.basicAuth(m.workerStateHandle))
	r.HandleFunc("/task/image/simple", m.basicAuth(m.newImageSimple))
	r.HandleFunc("/task/video/simple", m.basicAuth(m.newVideoSimpleHandle))
	r.HandleFunc("/task/video/vod", m.basicAuth(m.newVideoOnDemandHandle))
	r.HandleFunc("/task/video/probe", m.basicAuth(m.indexHandle))
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

	if err = t.ValidateRequest(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = m.mq.Push(&t, "video/vod")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	m.state.jobs[t.GetID()] = FullStatusIndicator{
		FailureMode: "IN-PROGRESS",
		Summary:     "Starting",
		Detail:      "VOD Job Sent to Proceessing",
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")

	rtn, err := json.MarshalIndent(TaskIdentification{
		State:  "encoding",
		TaskID: t.GetID(),
	}, "", "    ")

	if err != nil {
		http.Error(w,
			fmt.Sprintf("Encoding with Error - %v", err.Error()),
			http.StatusInternalServerError,
		)
	} else {
		w.Write(rtn)
	}

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

	if err = t.ValidateRequest(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = m.mq.Push(&t, "video/simple")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	m.state.jobs[t.GetID()] = FullStatusIndicator{
		FailureMode: "IN-PROGRESS",
		Summary:     "Starting",
		Detail:      "Simple Video Job Sent to Proceessing",
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")

	rtn, err := json.MarshalIndent(TaskIdentification{
		State:  "encoding",
		TaskID: t.GetID(),
	}, "", "    ")

	if err != nil {
		http.Error(w,
			fmt.Sprintf("Encoding with Error - %v", err.Error()),
			http.StatusInternalServerError,
		)
	} else {
		w.Write(rtn)
	}
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

// basicAuth wraps a handler requiring HTTP basic auth for it using the given
// username and password and the specified realm, which shouldn't contain quotes.
//
// Most web browser display a dialog with something like:
//
//    The website says: "<realm>"
//
// Which is really stupid so you may want to set the realm to a message rather than
// an actual realm.
func (m *Manager) basicAuth(handler http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		user, pass, ok := r.BasicAuth()

		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(m.user)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(m.pass)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="ystv vt"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}

		handler(w, r)
	}
}
