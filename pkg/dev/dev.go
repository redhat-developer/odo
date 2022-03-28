package dev

import (
	"context"
	"io"

	"github.com/redhat-developer/odo/pkg/kclient"

	"github.com/redhat-developer/odo/pkg/envinfo"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/util"
	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
	"github.com/redhat-developer/odo/pkg/watch"
	"k8s.io/klog/v2"
)

// this causes compilation to fail if DevClient struct doesn't implement Client interface
var _ Client = (*DevClient)(nil)

type DevClient struct {
	watchClient     watch.Client
	kuberneteClient kclient.ClientInterface
}

func NewDevClient(watchClient watch.Client, kubernetesClient kclient.ClientInterface) *DevClient {
	return &DevClient{
		watchClient:     watchClient,
		kuberneteClient: kubernetesClient,
	}
}

func (o *DevClient) Start(devfileObj parser.DevfileObj, platformContext kubernetes.KubernetesContext, ignorePaths []string, path string) error {
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
	}

	klog.V(4).Infoln("Creating inner-loop resources for the component")
	err = adapter.Push(pushParameters)
	if err != nil {
		return err
	}
	klog.V(4).Infoln("Successfully created inner-loop resources")
	return nil
}

func (o *DevClient) Watch(devfileObj parser.DevfileObj, path string, ignorePaths []string, out io.Writer, h Handler, cleanupDone chan bool, ctx context.Context) error {
	envSpecificInfo, err := envinfo.NewEnvSpecificInfo(path)
	if err != nil {
		return err
	}

	absIgnorePaths := util.GetAbsGlobExps(path, ignorePaths)

	watchParameters := watch.WatchParameters{
		Path:                path,
		ComponentName:       devfileObj.GetMetadataName(),
		ApplicationName:     "app",
		ExtChan:             make(chan bool),
		DevfileWatchHandler: h.RegenerateAdapterAndPush,
		EnvSpecificInfo:     envSpecificInfo,
		FileIgnores:         absIgnorePaths,
		InitialDevfileObj:   devfileObj,
	}

	return o.watchClient.WatchAndPush(out, watchParameters, cleanupDone, ctx)
}
