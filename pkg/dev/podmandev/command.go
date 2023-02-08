package podmandev

import (
	"context"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/component"
	envcontext "github.com/redhat-developer/odo/pkg/config/context"
	"github.com/redhat-developer/odo/pkg/devfile/image"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/platform"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

type commandHandler struct {
	ctx             context.Context
	fs              filesystem.Filesystem
	execClient      exec.Client
	platformClient  platform.Client
	componentExists bool
	podName         string
	appName         string
	componentName   string
}

var _ libdevfile.Handler = (*commandHandler)(nil)

func (a commandHandler) ApplyImage(img devfilev1.Component) error {
	return image.BuildPushSpecificImage(a.ctx, a.fs, img, envcontext.GetEnvConfig(a.ctx).PushImages)
}

func (a commandHandler) ApplyKubernetes(kubernetes devfilev1.Component) error {
	klog.V(4).Info("apply kubernetes commands are not implemented on podman")
	log.Warningf("Apply Kubernetes components are not supported on Podman. Skipping: %v.", kubernetes.Name)
	return nil
}

func (a commandHandler) ApplyOpenShift(openshift devfilev1.Component) error {
	klog.V(4).Info("apply OpenShift commands are not implemented on podman")
	log.Warningf("Apply OpenShift components are not supported on Podman. Skipping: %v.", openshift.Name)
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
