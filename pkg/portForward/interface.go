package portForward

import (
	"io"

	"github.com/devfile/library/pkg/devfile/parser"
)

type Client interface {
	StartPortForwarding(
		devFileObj parser.DevfileObj,
		componenentName string,
		randomPorts bool,
		errOut io.Writer,
	) error

	StopPortForwarding()
}
