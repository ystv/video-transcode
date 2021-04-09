package manager

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func (m *Manager) jobStateHandle(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	uuid := params["uuid"]

	jobState, ok := m.state.jobs[uuid]

	if !ok {
		http.Error(w,
			fmt.Sprintf("Job with UUID %s not found", uuid),
			http.StatusNotFound)
	}

	rtn, err := json.MarshalIndent(jobState, "", "    ")

	if err != nil {
		http.Error(w,
			"Error getting Job Status",
			http.StatusInternalServerError)
	} else {
		w.Write(rtn)
	}
}

func (m *Manager) workerStateHandle(w http.ResponseWriter, r *http.Request) {
	// TODO
	w.WriteHeader(http.StatusNotImplemented)
}
