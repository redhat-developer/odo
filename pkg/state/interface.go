package state

import (
	"context"

	"github.com/redhat-developer/odo/pkg/api"
)

type Client interface {
	// Init creates a devstate file for the process
	Init(ctx context.Context) error

	// SetForwardedPorts sets the forwarded ports in the state file and saves it to the file, updating the metadata
	SetForwardedPorts(ctx context.Context, fwPorts []api.ForwardedPort) error

	// GetForwardedPorts returns the ports forwarded by the current odo dev session
	GetForwardedPorts(ctx context.Context) ([]api.ForwardedPort, error)

	// SaveExit resets the state file to indicate odo is not running
	SaveExit(ctx context.Context) error

	// SetAPIServerPort sets the port where API server is listening in the state file and saves it to the file, updating the metadata
	SetAPIServerPort(ctx context.Context, port int) error

	// GetAPIServerPorts returns the port where the API servers are listening, possibly per platform.
	GetAPIServerPorts(ctx context.Context) ([]api.DevControlPlane, error)

	GetOrphanFiles(ctx context.Context) ([]string, error)
}
