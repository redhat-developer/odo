package common

import (
	"context"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/component"
	envcontext "github.com/redhat-developer/odo/pkg/config/context"
	"github.com/redhat-developer/odo/pkg/devfile/image"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/platform"
	"github.com/redhat-developer/odo/pkg/remotecmd"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

type RunHandler struct {
	FS              filesystem.Filesystem
	ExecClient      exec.Client
	AppName         string
	ComponentName   string
	Devfile         parser.DevfileObj
	PlatformClient  platform.Client
	ImageBackend    image.Backend
	Path            string
	ComponentExists bool
	PodName         string

	Ctx context.Context
}

var _ libdevfile.Handler = (*RunHandler)(nil)

func (a *RunHandler) ApplyImage(img devfilev1.Component) error {
	return image.BuildPushSpecificImage(a.Ctx, a.ImageBackend, a.FS, img, envcontext.GetEnvConfig(a.Ctx).PushImages)
}

func (a *RunHandler) ApplyKubernetes(kubernetes devfilev1.Component) error {
	switch platform := a.PlatformClient.(type) {
	case kclient.ClientInterface:
		return component.ApplyKubernetes(odolabels.ComponentDevMode, a.AppName, a.ComponentName, a.Devfile, kubernetes, platform, a.Path)
	default:
		klog.V(4).Info("apply kubernetes commands are not implemented on podman")
		log.Warningf("Apply Kubernetes components are not supported on Podman. Skipping: %v.", kubernetes.Name)
		return nil
	}
}

func (a *RunHandler) ApplyOpenShift(openshift devfilev1.Component) error {
	switch platform := a.PlatformClient.(type) {
	case kclient.ClientInterface:
		return component.ApplyKubernetes(odolabels.ComponentDevMode, a.AppName, a.ComponentName, a.Devfile, openshift, platform, a.Path)
	default:
		klog.V(4).Info("apply OpenShift commands are not implemented on podman")
		log.Warningf("Apply OpenShift components are not supported on Podman. Skipping: %v.", openshift.Name)
		return nil
	}
}

func (a *RunHandler) ExecuteNonTerminatingCommand(ctx context.Context, command devfilev1.Command) error {
	return component.ExecuteRunCommand(ctx, a.ExecClient, a.PlatformClient, command, a.ComponentExists, a.PodName, a.AppName, a.ComponentName)
}

func (a *RunHandler) ExecuteTerminatingCommand(ctx context.Context, command devfilev1.Command) error {
	return component.ExecuteRunCommand(ctx, a.ExecClient, a.PlatformClient, command, a.ComponentExists, a.PodName, a.AppName, a.ComponentName)
}

// IsRemoteProcessForCommandRunning returns true if the command is running
func (a *RunHandler) IsRemoteProcessForCommandRunning(ctx context.Context, command devfilev1.Command, podName string) (bool, error) {
	remoteProcess, err := remotecmd.NewKubeExecProcessHandler(a.ExecClient).GetProcessInfoForCommand(ctx, remotecmd.CommandDefinition{Id: command.Id}, podName, command.Exec.Component)
	if err != nil {
		return false, err
	}

	return remoteProcess.Status == remotecmd.Running, nil
}
