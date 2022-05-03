package state

import (
	"github.com/redhat-developer/odo/pkg/api"
)

type Content struct {
	// ForwardedPorts are the ports forwarded during odo dev session
	ForwardedPorts []api.ForwardedPort `json:"forwardedPorts"`
}
