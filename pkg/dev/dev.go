package dev

import (
	"context"
	"io"

	"github.com/redhat-developer/odo/pkg/binding"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/portForward"
	"github.com/redhat-developer/odo/pkg/preference"
	filesystem "github.com/redhat-developer/odo/pkg/testingutil/filesystem"

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
	filesystem        filesystem.Filesystem
}

var _ Client = (*DevClient)(nil)

func NewDevClient(
	kubernetesClient kclient.ClientInterface,
	prefClient preference.Client,
	portForwardClient portForward.Client,
	watchClient watch.Client,
	bindingClient binding.Client,
	filesystem filesystem.Filesystem,
) *DevClient {
	return &DevClient{
		kubernetesClient:  kubernetesClient,
		prefClient:        prefClient,
		portForwardClient: portForwardClient,
		watchClient:       watchClient,
		bindingClient:     bindingClient,
		filesystem:        filesystem,
	}
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
	handler Handler,
) error {
	klog.V(4).Infoln("Creating new adapter")

	adapter := k8sComponent.NewKubernetesAdapter(
		o.kubernetesClient, o.prefClient, o.portForwardClient, o.bindingClient,
		k8sComponent.AdapterContext{
			ComponentName: componentName,
			Context:       path,
			AppName:       "app",
			Devfile:       devfileObj,
			FS:            o.filesystem,
		})

	envSpecificInfo, err := envinfo.NewEnvSpecificInfo(path)
	if err != nil {
		return err
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
		return err
	}
	klog.V(4).Infoln("Successfully created inner-loop resources")

	watchParameters := watch.WatchParameters{
		DevfilePath:         devfilePath,
		Path:                path,
		ComponentName:       componentName,
		ApplicationName:     "app",
		DevfileWatchHandler: handler.RegenerateAdapterAndPush,
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
