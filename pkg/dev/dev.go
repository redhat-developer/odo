package dev

import (
	"io"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/watch"
	"k8s.io/klog/v2"
)

// this causes compilation to fail if DevClient struct doesn't implement Client interface
var _ Client = (*DevClient)(nil)

type DevClient struct {
	watchClient watch.Client
}

func NewDevClient(watchClient watch.Client) *DevClient {
	return &DevClient{
		watchClient: watchClient,
	}
}

// Start the resources in devfileObj on the platformContext. It then pushes the files in path to the container,
// and watches it for any changes. It prints all the logs/output to out.
func (o *DevClient) Start(devfileObj parser.DevfileObj, platformContext kubernetes.KubernetesContext, ignorePaths []string, path string, out io.Writer, h Handler) error {
	var err error

	var adapter common.ComponentAdapter
	klog.V(4).Infoln("Creating new adapter")
	adapter, err = adapters.NewComponentAdapter(devfileObj.GetMetadataName(), path, "app", devfileObj, platformContext)
	if err != nil {
		return err
	}

	var envSpecificInfo *envinfo.EnvSpecificInfo
	envSpecificInfo, err = envinfo.NewEnvSpecificInfo(path)
	if err != nil {
		return err
	}
	pushParameters := common.PushParameters{
		EnvSpecificInfo: *envSpecificInfo,
		Path:            path,
	}

	klog.V(4).Infoln("Creating inner-loop resources for the component")
	err = adapter.Push(pushParameters)
	if err != nil {
		return err
	}
	klog.V(4).Infoln("Successfully created inner-loop resourcs")
	log.Finfof(out, "\nYour application is now running on your cluster.")

	watchParameters := watch.WatchParameters{
		Path:                path,
		ComponentName:       devfileObj.GetMetadataName(),
		ApplicationName:     "app",
		ExtChan:             make(chan bool),
		DevfileWatchHandler: h.RegenerateAdapterAndPush,
		EnvSpecificInfo:     envSpecificInfo,
		FileIgnores:         ignorePaths,
	}

	return o.watchClient.WatchAndPush(out, watchParameters)
}

// Cleanup cleans the resources created by Push
func (o *DevClient) Cleanup() error {
	var err error
	return err
}
