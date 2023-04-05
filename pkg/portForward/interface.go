package portForward

import (
	"context"
	"io"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"

	"github.com/redhat-developer/odo/pkg/api"
)

type Client interface {
	// StartPortForwarding starts port forwarding for the endpoints defined in the containers of the devfile
	// componentName indicates the name of component in the Devfile
	// randomPorts indicates to affect random ports, instead of stable ports starting at 20001
	// output will be written to errOut writer
	// definedPorts allows callers to explicitly define the mapping they want to set.
	StartPortForwarding(
		ctx context.Context,
		devFileObj parser.DevfileObj,
		componentName string,
		debug bool,
		randomPorts bool,
		out io.Writer,
		errOut io.Writer,
		definedPorts []api.ForwardedPort,
	) error

	// StopPortForwarding stops the port forwarding for the specified component.
	StopPortForwarding(componentName string)

	// GetForwardedPorts returns the list of ports for each container currently forwarded.
	GetForwardedPorts() map[string][]v1alpha2.Endpoint
}
