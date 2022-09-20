package dev

import (
	"context"
	"io"

	"github.com/redhat-developer/odo/pkg/binding"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/portForward"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"

	"github.com/devfile/library/pkg/devfile/parser"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	k8sComponent "github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/component"
	"github.com/redhat-developer/odo/pkg/watch"
)

type DevClient struct {
	kubernetesClient  kclient.ClientInterface
	prefClient        preference.Client
	portForwardClient portForward.Client
	watchClient       watch.Client
	bindingClient     binding.Client
}

var _ Client = (*DevClient)(nil)

func NewDevClient(
	kubernetesClient kclient.ClientInterface,
	prefClient preference.Client,
	portForwardClient portForward.Client,
	watchClient watch.Client,
	bindingClient binding.Client,
) *DevClient {
	return &DevClient{
		kubernetesClient:  kubernetesClient,
		prefClient:        prefClient,
		portForwardClient: portForwardClient,
		watchClient:       watchClient,
		bindingClient:     bindingClient,
	}
}

func (o *DevClient) Start(
	devfileObj parser.DevfileObj,
	componentName string,
	ignorePaths []string,
	path string,
	debug bool,
	buildCommand string,
	runCommand string,
	randomPorts bool,
	errOut io.Writer,
	fs filesystem.Filesystem,
) (watch.ComponentStatus, error) {
	klog.V(4).Infoln("Creating new adapter")

	adapter := k8sComponent.NewKubernetesAdapter(
		o.kubernetesClient, o.prefClient, o.portForwardClient, o.bindingClient,
		k8sComponent.AdapterContext{
			ComponentName: componentName,
			Context:       path,
			AppName:       "app",
			Devfile:       devfileObj,
			FS:            fs,
		})

	envSpecificInfo, err := envinfo.NewEnvSpecificInfo(path)
	if err != nil {
		return watch.ComponentStatus{}, err
	}

	pushParameters := adapters.PushParameters{
		EnvSpecificInfo: *envSpecificInfo,
		Path:            path,
		IgnoredFiles:    ignorePaths,
		Debug:           debug,
		DevfileBuildCmd: buildCommand,
		DevfileRunCmd:   runCommand,
		RandomPorts:     randomPorts,
		ErrOut:          errOut,
	}

	klog.V(4).Infoln("Creating inner-loop resources for the component")
	componentStatus := watch.ComponentStatus{}
	err = adapter.Push(pushParameters, &componentStatus)
	if err != nil {
		return watch.ComponentStatus{}, err
	}
	klog.V(4).Infoln("Successfully created inner-loop resources")
	return componentStatus, nil
}

func (o *DevClient) Watch(
	devfilePath string,
	devfileObj parser.DevfileObj,
	componentName string,
	path string,
	ignorePaths []string,
	out io.Writer,
	h Handler,
	ctx context.Context,
	debug bool,
	buildCommand string,
	runCommand string,
	variables map[string]string,
	randomPorts bool,
	watchFiles bool,
	errOut io.Writer,
	componentStatus watch.ComponentStatus,
) error {
	envSpecificInfo, err := envinfo.NewEnvSpecificInfo(path)
	if err != nil {
		return err
	}

	watchParameters := watch.WatchParameters{
		DevfilePath:         devfilePath,
		Path:                path,
		ComponentName:       componentName,
		ApplicationName:     "app",
		DevfileWatchHandler: h.RegenerateAdapterAndPush,
		EnvSpecificInfo:     envSpecificInfo,
		FileIgnores:         ignorePaths,
		InitialDevfileObj:   devfileObj,
		Debug:               debug,
		DevfileBuildCmd:     buildCommand,
		DevfileRunCmd:       runCommand,
		Variables:           variables,
		RandomPorts:         randomPorts,
		WatchFiles:          watchFiles,
		ErrOut:              errOut,
	}

	return o.watchClient.WatchAndPush(out, watchParameters, ctx, componentStatus)
}
