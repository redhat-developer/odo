package podmandev

import (
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/platform"
)

type commandHandler struct {
	execClient      exec.Client
	platformClient  platform.Client
	componentExists bool
	podName         string
	appName         string
	componentName   string
}

var _ libdevfile.Handler = (*commandHandler)(nil)

func (a commandHandler) ApplyImage(img devfilev1.Component) error {
	// Not implemented
	return nil
}

func (a commandHandler) ApplyKubernetes(kubernetes devfilev1.Component) error {
	// Not implemented
	return nil
}

func (a commandHandler) Execute(devfileCmd devfilev1.Command) error {
	return component.ExecuteRunCommand(
		a.execClient,
		a.platformClient,
		devfileCmd,
		a.componentExists,
		a.podName,
		a.appName,
		a.componentName,
	)
}
