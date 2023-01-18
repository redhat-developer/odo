package portForward

import (
	"io"

	"github.com/devfile/library/v2/pkg/devfile/parser"
)

type Client interface {
	// StartPortForwarding starts port forwarding for the endpoints defined in the containers of the devfile
	// componentName indicates the name of component in the Devfile
	// randomPorts indicates to affect random ports, instead of stable ports starting at 40001
	// output will be written to errOut writer
	StartPortForwarding(
		devFileObj parser.DevfileObj,
		componenentName string,
		debug bool,
		randomPorts bool,
		errOut io.Writer,
	) error

	// StopPortForwarding stops the port forwarding
	StopPortForwarding()

	// GetForwardedPorts returns the list of ports for each containers currently forwarded
	GetForwardedPorts() map[string][]int

	// GetPortsToForward returns the endpoints to forward from the Devfile.
	// Debug ports will be included only if includeDebug is true.
	GetPortsToForward(devFileObj parser.DevfileObj, includeDebug bool) (map[string][]int, error)
}
