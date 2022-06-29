package dev

import (
	"context"
	"io"

	"github.com/redhat-developer/odo/pkg/envinfo"

	"github.com/devfile/library/pkg/devfile/parser"
	"k8s.io/klog/v2"

	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
	"github.com/redhat-developer/odo/pkg/watch"
)

type DevClient struct {
	watchClient watch.Client
}

var _ Client = (*DevClient)(nil)

func NewDevClient(watchClient watch.Client) *DevClient {
	return &DevClient{
		watchClient: watchClient,
	}
}

func (o *DevClient) Start(
	devfileObj parser.DevfileObj,
	platformContext kubernetes.KubernetesContext,
	ignorePaths []string,
	path string,
	debug bool,
	buildCommand string,
	runCommand string,
) error {
	klog.V(4).Infoln("Creating new adapter")
	adapter, err := adapters.NewComponentAdapter(devfileObj.GetMetadataName(), path, "app", devfileObj, platformContext)
	if err != nil {
		return err
	}

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
