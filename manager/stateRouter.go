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

	jobState, ok := m.state.Jobs[uuid]

	if !ok {
		http.Error(w,
			fmt.Sprintf("Job with UUID %s not found", uuid),
			http.StatusNotFound)
		return
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
	params := mux.Vars(r)
	uuid := params["uuid"]

	workerState, ok := m.state.Workers[uuid]

	if !ok {
		http.Error(w,
			fmt.Sprintf("Worker with UUID %s not found", uuid),
			http.StatusNotFound)
		return
	}

	rtn, err := json.MarshalIndent(workerState, "", "    ")

	if err != nil {
		http.Error(w,
			"Error getting Worker Status",
			http.StatusInternalServerError)
	} else {
		w.Write(rtn)
	}
}

func (m *Manager) allWorkersHandler(w http.ResponseWriter, r *http.Request) {
	rtn, err := json.MarshalIndent(m.state.Workers, "", "    ")
	if err != nil {
		http.Error(w,
			"Error getting all worker statuses",
			http.StatusInternalServerError)
	} else {
		w.Write(rtn)
	}
}
