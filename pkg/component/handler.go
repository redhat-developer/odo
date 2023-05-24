package component

import (
	"context"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"k8s.io/klog"

	envcontext "github.com/redhat-developer/odo/pkg/config/context"
	"github.com/redhat-developer/odo/pkg/configAutomount"
	"github.com/redhat-developer/odo/pkg/devfile/image"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/platform"
	"github.com/redhat-developer/odo/pkg/remotecmd"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

type runHandler struct {
	ctx                   context.Context
	platformClient        platform.Client
	execClient            exec.Client
	configAutomountClient configAutomount.Client
	podName               string
	ComponentExists       bool
	containersRunning     []string
	msg                   string

	fs           filesystem.Filesystem
	imageBackend image.Backend

	devfile parser.DevfileObj
	path    string
}

var _ libdevfile.Handler = (*runHandler)(nil)

func NewRunHandler(
	ctx context.Context,
	platformClient platform.Client,
	execClient exec.Client,
	configAutomountClient configAutomount.Client,
	podName string,
	componentExists bool,
	containersRunning []string,
	msg string,

	// For building images
	fs filesystem.Filesystem,
	imageBackend image.Backend,

	// For apply Kubernetes / Openshift
	devfile parser.DevfileObj,
	path string,

) *runHandler {
	return &runHandler{
		ctx:                   ctx,
		platformClient:        platformClient,
		execClient:            execClient,
		configAutomountClient: configAutomountClient,
		podName:               podName,
		ComponentExists:       componentExists,
		containersRunning:     containersRunning,
		msg:                   msg,

		fs:           fs,
		imageBackend: imageBackend,

		devfile: devfile,
		path:    path,
	}
}

func (a *runHandler) ApplyImage(img devfilev1.Component) error {
	return image.BuildPushSpecificImage(a.ctx, a.imageBackend, a.fs, img, envcontext.GetEnvConfig(a.ctx).PushImages)
}

func (a *runHandler) ApplyKubernetes(kubernetes devfilev1.Component, kind v1alpha2.CommandGroupKind) error {
	var (
		componentName = odocontext.GetComponentName(a.ctx)
		appName       = odocontext.GetApplication(a.ctx)
	)
	mode := odolabels.ComponentDevMode
	if kind == v1alpha2.DeployCommandGroupKind {
		mode = odolabels.ComponentDeployMode
	}
	switch platform := a.platformClient.(type) {
	case kclient.ClientInterface:
		return ApplyKubernetes(mode, appName, componentName, a.devfile, kubernetes, platform, a.path)
	default:
		klog.V(4).Info("apply kubernetes commands are not implemented on podman")
		log.Warningf("Apply Kubernetes components are not supported on Podman. Skipping: %v.", kubernetes.Name)
		return nil
	}
}

func (a *runHandler) ApplyOpenShift(openshift devfilev1.Component, kind v1alpha2.CommandGroupKind) error {
	var (
		componentName = odocontext.GetComponentName(a.ctx)
		appName       = odocontext.GetApplication(a.ctx)
	)
	mode := odolabels.ComponentDevMode
	if kind == v1alpha2.DeployCommandGroupKind {
		mode = odolabels.ComponentDeployMode
	}
	switch platform := a.platformClient.(type) {
	case kclient.ClientInterface:
		return ApplyKubernetes(mode, appName, componentName, a.devfile, openshift, platform, a.path)
	default:
		klog.V(4).Info("apply OpenShift commands are not implemented on podman")
		log.Warningf("Apply OpenShift components are not supported on Podman. Skipping: %v.", openshift.Name)
		return nil
	}
}

func (a *runHandler) ExecuteNonTerminatingCommand(ctx context.Context, command devfilev1.Command) error {
	var (
		componentName = odocontext.GetComponentName(a.ctx)
		appName       = odocontext.GetApplication(a.ctx)
	)
	if isContainerRunning(command.Exec.Component, a.containersRunning) {
		return ExecuteRunCommand(ctx, a.execClient, a.platformClient, command, a.ComponentExists, a.podName, appName, componentName)
	}
	switch platform := a.platformClient.(type) {
	case kclient.ClientInterface:
		return ExecuteInNewContainer(ctx, platform, a.configAutomountClient, a.devfile, componentName, appName, command)
	default:
		klog.V(4).Info("executing a command in a new container is not implemented on podman")
		log.Warningf("executing a command in a new container is not implemented on podman. Skipping: %v.", command.Id)
		return nil
	}
}

func (a *runHandler) ExecuteTerminatingCommand(ctx context.Context, command devfilev1.Command) error {
	var (
		componentName = odocontext.GetComponentName(a.ctx)
		appName       = odocontext.GetApplication(a.ctx)
	)
	if isContainerRunning(command.Exec.Component, a.containersRunning) {
		return ExecuteTerminatingCommand(ctx, a.execClient, a.platformClient, command, a.ComponentExists, a.podName, appName, componentName, a.msg, false)
	}
	switch platform := a.platformClient.(type) {
	case kclient.ClientInterface:
		return ExecuteInNewContainer(ctx, platform, a.configAutomountClient, a.devfile, componentName, appName, command)
	default:
		klog.V(4).Info("executing a command in a new container is not implemented on podman")
		log.Warningf("executing a command in a new container is not implemented on podman. Skipping: %v.", command.Id)
		return nil
	}
}

// IsRemoteProcessForCommandRunning returns true if the command is running
func (a *runHandler) IsRemoteProcessForCommandRunning(ctx context.Context, command devfilev1.Command, podName string) (bool, error) {
	remoteProcess, err := remotecmd.NewKubeExecProcessHandler(a.execClient).GetProcessInfoForCommand(ctx, remotecmd.CommandDefinition{Id: command.Id}, podName, command.Exec.Component)
	if err != nil {
		return false, err
	}

	return remoteProcess.Status == remotecmd.Running, nil
}

func isContainerRunning(container string, containers []string) bool {
	for _, cnt := range containers {
		if container == cnt {
			return true
		}
	}
	return false
}
