package dev

import (
	devfilev2 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/watch"
	"io"
)

// this causes compilation to fail if DevClient struct doesn't implement Client interface
var _ Client = (*DevClient)(nil)

type DevClient struct {
	client kclient.ClientInterface
	// devfileObj is stored for Cleanup; ideally populated by Start method
	//devfileObj parser.DevfileObj
}

func NewDevClient(client kclient.ClientInterface) *DevClient {
	return &DevClient{
		client: client,
	}
}

// getComponents returns a slice of components to be started for inner loop
func getComponents() (devfilev2.Component, error) {
	var components devfilev2.Component
	var err error
	return components, err
}

// Start the resources in devfileObj on the platformContext. It then pushes the files in path to the container,
// and watches it for any changes. It prints all the logs/output to out.
func (o *DevClient) Start(devfileObj parser.DevfileObj, platformContext kubernetes.KubernetesContext, path string, out io.Writer) error {
	var err error

	var adapter common.ComponentAdapter
	adapter, err = adapters.NewComponentAdapter(devfileObj.GetMetadataName(), path, "app", devfileObj, platformContext)
	if err != nil {
		return err
	}

	// store the devfileObj so that we can reuse it in Cleanup
	// o.devfileObj = devfileObj
	var envSpecificInfo *envinfo.EnvSpecificInfo
	envSpecificInfo, err = envinfo.NewEnvSpecificInfo(path)
	pushParameters := common.PushParameters{
		EnvSpecificInfo: *envSpecificInfo,
		Path:            path,
	}

	err = adapter.Push(pushParameters)
	if err != nil {
		return err
	}

	watchParameters := watch.WatchParameters{
		Path:                path,
		ComponentName:       devfileObj.GetMetadataName(),
		ApplicationName:     "app",
		ExtChan:             make(chan bool),
		DevfileWatchHandler: regenerateAdapterAndPush,
		EnvSpecificInfo:     envSpecificInfo,
	}

	err = watch.WatchAndPush(o.client, out, watchParameters)
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
		return errors.Wrapf(err, "unable to generate component from watch parameters")
	}

	err = adapter.Push(pushParams)
	if err != nil {
		return errors.Wrapf(err, "watch command was unable to push component")
	}

	return err
}

func regenerateComponentAdapterFromWatchParams(parameters watch.WatchParameters) (common.ComponentAdapter, error) {

	// Parse devfile and validate
	devObj, err := devfile.ParseAndValidateFromFile("./devfile.yaml")
	if err != nil {
		return nil, err
	}

	platformContext := kubernetes.KubernetesContext{
		// TODO: find a better way, or get RID of KubernetesContext
		Namespace: "myproject",
	}

	return adapters.NewComponentAdapter(parameters.ComponentName, parameters.Path, parameters.ApplicationName, devObj, platformContext)

}
