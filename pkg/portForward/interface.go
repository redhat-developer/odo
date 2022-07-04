package portForward

import (
	"io"

	"github.com/devfile/library/pkg/devfile/parser"
)

type Client interface {
	// StartPortForwarding starts port forwarding for the endpoints defined in the containers of the devfile
	// componentName indicates the name of component in the Devfile
	// randomPorts indicates to affect random ports, instead of stable ports starting at 40001
	// output will be written to errOut writer
	StartPortForwarding(
		devFileObj parser.DevfileObj,
		componenentName string,
		randomPorts bool,
		errOut io.Writer,
	) error

	// StopPortForwarding stops the port forwarding
	StopPortForwarding()
}
