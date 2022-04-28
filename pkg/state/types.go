package state

import (
	"github.com/redhat-developer/odo/pkg/api"
)

type Content struct {
	// Timestamp is the number of seconds from epoch at which the state were saved
	Timestamp int64 `json:"timestamp"`
	// PID is the pid of the running odo process, 0 if odo is not running
	PID int `json:"pid"`
	// ForwardedPorts are the ports forwarded during odo dev session
	ForwardedPorts []api.ForwardedPort `json:"forwardedPorts"`
}
