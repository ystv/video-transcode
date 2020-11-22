package main

import (
	"encoding/json"
	"net/http"

	"github.com/ystv/video-transcode/event"
)

// IndexHandle just shows it's alive, could have metrics?
func IndexHandle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("vt"))
}

// NewVODHandle will download file from CDN to local
// transcode, upload and cleanup
func NewVODHandle(w http.ResponseWriter, r *http.Request) {

}

// NewLiveHandle will stream file to ffmpeg, optional
// transcode and send to new endpoint
func NewLiveHandle(w http.ResponseWriter, r *http.Request) {
	t := event.Task{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = emitter.Push(t, "live")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("encoding"))
}
