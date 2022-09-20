package podmandev

import (
	"context"
	"fmt"
	"io"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/dev"
)

type DevClient struct {
}

var _ dev.Client = (*DevClient)(nil)

func NewDevClient() *DevClient {
	return &DevClient{}
}

func (o *DevClient) Start(
	ctx context.Context,
	devfileObj parser.DevfileObj,
	componentName string,
	path string,
	devfilePath string,
	ignorePaths []string,
	debug bool,
	buildCommand string,
	runCommand string,
	randomPorts bool,
	watchFiles bool,
	variables map[string]string,
	out io.Writer,
	errOut io.Writer,
) error {
	fmt.Printf("Deploying using Podman")
	return nil
}
