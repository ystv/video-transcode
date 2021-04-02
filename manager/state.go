package manager

// JobStatus defines methods for any status a job may be in,
// whether they be success states, in-progress states or
// failure states.
// TODO: Decisions about length of time success/failure states
// are kept for. Possibly a day with detailed information, and
// a week in summary. Some goroutines can Henry-Hoover this all up
type JobStatus interface {
	Get() string
	Failure() bool
	DetailedStatus() string
}

// StateHandler is the central place for the systems status
// for access over HTTP by users
type stateHandler struct {
	// TODO: Implement Worker Statuses Here Too
	jobs map[string]JobStatus
}

func newStateHandler() *stateHandler {
	jobs := make(map[string]JobStatus)
	return &stateHandler{
		jobs: jobs,
	}
}

// TaskIdentification is for initially informing the user
// of their job starting and its given ID for later
// checking
type TaskIdentification struct {
	State  string `json:"state"`
	TaskID string `json:"taskID"`
}
