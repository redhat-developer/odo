package state

import "github.com/redhat-developer/odo/pkg/api"

type Client interface {
	// SetForwardedPorts sets the forwarded ports in the state file and saves it to the file, updating the metadata
	SetForwardedPorts(fwPorts []api.ForwardedPort) error

	// GetForwardedPorts returns the ports forwarded by the current odo dev session
	GetForwardedPorts() ([]api.ForwardedPort, error)

	// SaveExit resets the state file to indicate odo is not running
	SaveExit() error
}
