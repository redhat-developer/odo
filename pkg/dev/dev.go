package dev

import (
	"fmt"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/watch"
	"io"
	"k8s.io/klog/v2"
)

// this causes compilation to fail if DevClient struct doesn't implement Client interface
var _ Client = (*DevClient)(nil)

type DevClient struct {
	kubernetesClient kclient.ClientInterface
	watchClient      watch.Client
}

func NewDevClient(kubernetesClient kclient.ClientInterface, watchClient watch.Client) *DevClient {
	return &DevClient{
		kubernetesClient: kubernetesClient,
		watchClient:      watchClient,
	}
}

// Start the resources in devfileObj on the platformContext. It then pushes the files in path to the container,
// and watches it for any changes. It prints all the logs/output to out.
func (o *DevClient) Start(devfileObj parser.DevfileObj, platformContext kubernetes.KubernetesContext, ignorePaths []string, path string, out io.Writer) error {
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
	fmt.Fprintf(out, "\nYour application is running on cluster.\n ")

	watchParameters := watch.WatchParameters{
		Path:                path,
		ComponentName:       devfileObj.GetMetadataName(),
		ApplicationName:     "app",
		ExtChan:             make(chan bool),
		DevfileWatchHandler: regenerateAdapterAndPush,
		EnvSpecificInfo:     envSpecificInfo,
		FileIgnores:         ignorePaths,
	}

	err = o.watchClient.WatchAndPush(out, watchParameters)
	if err != nil {
		return err
	}
	return err
}

// Cleanup cleans the resources created by Push
func (o *DevClient) Cleanup() error {
	var err error
	return err
}

func regenerateAdapterAndPush(pushParams common.PushParameters, watchParams watch.WatchParameters) error {
	var adapter common.ComponentAdapter

	adapter, err := regenerateComponentAdapterFromWatchParams(watchParams)
	if err != nil {
		return fmt.Errorf("unable to generate component from watch parameters: %w", err)
	}

	err = adapter.Push(pushParams)
	if err != nil {
		return fmt.Errorf("watch command was unable to push component: %w", err)
	}

	return err
}

func regenerateComponentAdapterFromWatchParams(parameters watch.WatchParameters) (common.ComponentAdapter, error) {

	// Parse devfile and validate. Path is hard coded because odo expects devfile.yaml to be present in the pwd/cwd.
	devObj, err := devfile.ParseAndValidateFromFile("./devfile.yaml")
	if err != nil {
		return nil, err
	}

	platformContext := kubernetes.KubernetesContext{
		Namespace: parameters.EnvSpecificInfo.GetNamespace(),
	}

	return adapters.NewComponentAdapter(parameters.ComponentName, parameters.Path, parameters.ApplicationName, devObj, platformContext)

}
