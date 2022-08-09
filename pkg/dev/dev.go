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
	"k8s.io/klog/v2"

	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/component"
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
	namespace string,
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
	adapter := component.NewKubernetesAdapter(
		o.kubernetesClient, o.prefClient, o.portForwardClient, o.bindingClient,
		component.AdapterContext{
			ComponentName: devfileObj.GetMetadataName(),
			Context:       path,
			AppName:       "app",
			Devfile:       devfileObj,
			FS:            fs,
		},
		namespace)

	envSpecificInfo, err := envinfo.NewEnvSpecificInfo(path)
	if err != nil {
		return watch.ComponentStatus{}, err
	}

	pushParameters := adapters.PushParameters{
		EnvSpecificInfo: *envSpecificInfo,
		DebugPort:       envSpecificInfo.GetDebugPort(),
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
		ComponentName:       devfileObj.GetMetadataName(),
		ApplicationName:     "app",
		DevfileWatchHandler: h.RegenerateAdapterAndPush,
		EnvSpecificInfo:     envSpecificInfo,
		FileIgnores:         ignorePaths,
		InitialDevfileObj:   devfileObj,
		Debug:               debug,
		DevfileBuildCmd:     buildCommand,
		DevfileRunCmd:       runCommand,
		DebugPort:           envSpecificInfo.GetDebugPort(),
		Variables:           variables,
		RandomPorts:         randomPorts,
		WatchFiles:          watchFiles,
		ErrOut:              errOut,
	}

	return o.watchClient.WatchAndPush(out, watchParameters, ctx, componentStatus)
}
