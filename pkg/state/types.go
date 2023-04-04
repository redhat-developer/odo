package state

import (
	"github.com/redhat-developer/odo/pkg/api"
)

type Content struct {
	// PID is the ID of the process to which the state belongs
	PID int `json:"pid"`
	// ForwardedPorts are the ports forwarded during odo dev session
	ForwardedPorts []api.ForwardedPort `json:"forwardedPorts"`
}
