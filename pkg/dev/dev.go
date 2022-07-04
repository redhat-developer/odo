package dev

import (
	"context"
	"io"

	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/portForward"
	"github.com/redhat-developer/odo/pkg/preference"

	"github.com/devfile/library/pkg/devfile/parser"
	"k8s.io/klog/v2"

	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/component"
	"github.com/redhat-developer/odo/pkg/watch"
)

type DevClient struct {
	kubernetesClient  kclient.ClientInterface
	prefClient        preference.Client
	portForwardClient portForward.Client
	watchClient       watch.Client
}

var _ Client = (*DevClient)(nil)

func NewDevClient(
	kubernetesClient kclient.ClientInterface,
	prefClient preference.Client,
	portForwardClient portForward.Client,
	watchClient watch.Client,
) *DevClient {
	return &DevClient{
		kubernetesClient:  kubernetesClient,
		prefClient:        prefClient,
		portForwardClient: portForwardClient,
		watchClient:       watchClient,
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
) error {
	klog.V(4).Infoln("Creating new adapter")
	adapter := component.NewKubernetesAdapter(
		o.kubernetesClient, o.prefClient, o.portForwardClient,
		component.AdapterContext{
			ComponentName: devfileObj.GetMetadataName(),
			Context:       path,
			AppName:       "app",
			Devfile:       devfileObj,
		},
		namespace, randomPorts, errOut)

	envSpecificInfo, err := envinfo.NewEnvSpecificInfo(path)
	if err != nil {
		return err
	}

	pushParameters := common.PushParameters{
		EnvSpecificInfo: *envSpecificInfo,
		DebugPort:       envSpecificInfo.GetDebugPort(),
		Path:            path,
		IgnoredFiles:    ignorePaths,
		Debug:           debug,
		DevfileBuildCmd: buildCommand,
		DevfileRunCmd:   runCommand,
	}

	klog.V(4).Infoln("Creating inner-loop resources for the component")
	err = adapter.Push(pushParameters)
	if err != nil {
		return err
	}
	klog.V(4).Infoln("Successfully created inner-loop resources")
	return nil
}

func (o *DevClient) Watch(
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
) error {
	envSpecificInfo, err := envinfo.NewEnvSpecificInfo(path)
	if err != nil {
		return err
	}

	watchParameters := watch.WatchParameters{
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
	}

	return o.watchClient.WatchAndPush(out, watchParameters, ctx)
}
