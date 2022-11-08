package component

import (
	"context"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/devfile/image"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/remotecmd"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

type runHandler struct {
	fs              filesystem.Filesystem
	execClient      exec.Client
	appName         string
	componentName   string
	devfile         parser.DevfileObj
	kubeClient      kclient.ClientInterface
	path            string
	componentExists bool
	podName         string

	ctx context.Context
}

var _ libdevfile.Handler = (*runHandler)(nil)

func (a *runHandler) ApplyImage(img devfilev1.Component) error {
	return image.BuildPushSpecificImage(a.ctx, a.fs, a.path, img, true)
}

func (a *runHandler) ApplyKubernetes(kubernetes devfilev1.Component) error {
	return component.ApplyKubernetes(odolabels.ComponentDevMode, a.appName, a.componentName, a.devfile, kubernetes, a.kubeClient, a.path)
}

func (a *runHandler) Execute(devfileCmd devfilev1.Command) error {
	return component.ExecuteRunCommand(a.execClient, a.kubeClient, devfileCmd, a.componentExists, a.podName, a.appName, a.componentName)

}

// IsRemoteProcessForCommandRunning returns true if the command is running
func (a *runHandler) IsRemoteProcessForCommandRunning(command devfilev1.Command, podName string) (bool, error) {
	remoteProcess, err := remotecmd.NewKubeExecProcessHandler(a.execClient).GetProcessInfoForCommand(
		remotecmd.CommandDefinition{Id: command.Id}, podName, command.Exec.Component)
	if err != nil {
		return false, err
	}

	return remoteProcess.Status == remotecmd.Running, nil
}
