package manager

import (
	"github.com/ystv/video-transcode/event"
)

// Manager provides workers with jobs and offers REST
// endpoints for 3rd party applications
type Manager struct {
	user string
	pass string
	mq   *event.Producer
}

// New creates a new manager
func New(mq *event.Producer, user, pass string) *Manager {
	return &Manager{mq: mq, user: user, pass: pass}
}
