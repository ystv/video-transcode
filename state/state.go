package state

import "time"

// So, in general, all of this stuff could be moved to a new package.
// Probably a good idea, to create some nice code layout for all
// the neccesary methods and stuff

// JobStatus defines methods for any status a job may be in,
// whether they be success states, in-progress states or
// failure states.
// TODO: Decisions about length of time success/failure states
// are kept for. Possibly a day with detailed information, and
// a week in summary. Some goroutines can Henry-Hoover this all up
type JobStatus interface {
	Get() string
	GetUUID() string
	Failure() bool
	DetailedStatus() string
}

// WorkerStatus is the data related to an individual
// worker, that we can monitor
type WorkerStatus struct {
	JobsCount int `json:"jobsCount"`
}

// Busy returns whether the worker has any jobs running
func (w WorkerStatus) Busy() bool {
	return w.JobsCount != 0
}

func (w *WorkerStatus) StartJob() {
	w.JobsCount++
}

func (w *WorkerStatus) EndJob() {
	w.JobsCount--
}

// StateHandler is the central place for the systems status
// for access over HTTP by users
type StateHandler struct {
	// TODO: Implement Worker Statuses Here Too
	Jobs    map[string]JobStatus
	Workers map[string]*WorkerStatus
}

func NewStateHandler() *StateHandler {
	newSH := &StateHandler{
		Jobs:    make(map[string]JobStatus),
		Workers: make(map[string]*WorkerStatus),
	}
	go newSH.Tidier() // Henry Hoover
	return newSH
}

// TaskIdentification is for initially informing the user
// of their job starting and its given ID for later
// checking
type TaskIdentification struct {
	State  string `json:"state"`
	TaskID string `json:"taskID"`
}

// FullStatusIndicator is a generic structure for giving
// all available information about a job state, for use
// before it gets Henry-Hoovered.
type FullStatusIndicator struct {
	JobID       string    `json:"jobID"`
	FailureMode string    `json:"failureMode"`
	Summary     string    `json:"summary"`
	Detail      string    `json:"detail"`
	Time        time.Time `json:"time"`
}

// Get returns the job status summary.
func (fsi FullStatusIndicator) Get() string {
	return fsi.Summary
}

func (fsi FullStatusIndicator) GetUUID() string {
	return fsi.JobID
}

// Failure returns a bool for whether the job has failed
// The default case is that it has failed, cause
// it should always have a FailureMode
func (fsi FullStatusIndicator) Failure() bool {
	switch fsi.FailureMode {
	case "IN-PROGRESS":
		return false
	case "COMPLETED-OK":
		return false
	case "FAILED":
		return true
	default:
		// If it doesn't meet a failure state,
		// somethings done a bad
		return true
	}
}

// DetailedStatus returns a detailed form of the job
// status. This only lasts until the first Henry-Hoover
// process (see top)
func (fsi FullStatusIndicator) DetailedStatus() string {
	return fsi.Detail
}

// ShortStatusIndicator is a generic structure for giving
// a less detailed form of the job status.
// TODO: Decide about ExpiredTime. Its the expiration
// of the previous FullStatusIndicator. Should that have
// a future expiry time, and should the SSI also have
// one? Also, better naming of the field.
type ShortStatusIndicator struct {
	JobID           string    `json:"jobID"`
	FailureMode     string    `json:"failureMode"`
	Summary         string    `json:"summary"`
	Time            time.Time `json:"time"`
	FullExpiredTime time.Time `json:"fullExpiredTime"`
}

// Get returns the job status summary.
func (ssi ShortStatusIndicator) Get() string {
	return ssi.Summary
}

func (ssi ShortStatusIndicator) GetUUID() string {
	return ssi.JobID
}

// Failure returns a bool for whether the job has failed
// The default case is that it has failed, cause
// it should always have a FailureMode
func (ssi ShortStatusIndicator) Failure() bool {
	switch ssi.FailureMode {
	case "IN-PROGRESS":
		return false
	case "COMPLETED-OK":
		return false
	case "FAILED":
		return true
	default:
		// If it doesn't meet a failure state,
		// somethings done a bad
		return true
	}
}

// DetailedStatus just says that the detailed status
// has expired.
// TODO - Do we include timing information here?
func (ssi ShortStatusIndicator) DetailedStatus() string {
	return "Detailed Description Expired"
}

type WorkerStatusUpdate struct {
	State    string `json:"state"`
	WorkerID string `json:"workerID"`
}

type StatusUpdate struct {
	Header string      `json:"header"`
	Body   interface{} `json:"body"`
}
