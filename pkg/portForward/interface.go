package portForward

import (
	"io"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
)

type Client interface {
	// StartPortForwarding starts port forwarding for the endpoints defined in the containers of the devfile
	// componentName indicates the name of component in the Devfile
	// randomPorts indicates to affect random ports, instead of stable ports starting at 20001
	// output will be written to errOut writer
	StartPortForwarding(
		devFileObj parser.DevfileObj,
		componentName string,
		debug bool,
		randomPorts bool,
		errOut io.Writer,
	) error

	// StopPortForwarding stops the port forwarding
	StopPortForwarding()

	// GetForwardedPorts returns the list of ports for each container currently forwarded.
	GetForwardedPorts() map[string][]v1alpha2.Endpoint
}
