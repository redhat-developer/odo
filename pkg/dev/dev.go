package dev

import (
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/pkg/errors"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
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
	adapter, err = adapters.NewComponentAdapter(devfileObj.GetMetadataName(), path, "app", devfileObj, platformContext)
	if err != nil {
		return err
	}

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
		FileIgnores:         ignorePaths,
	}

	err = o.watchClient.WatchAndPush(out, watchParameters)
	if err != nil {
		return err
	}
	return err
}

// Cleanup cleans the resources created by Start by deleting the Kubernetes Deployment matching the .metadata.name
// in the devfileObj. It silently fails if it can't find a matching Deployment because it's possible that a Deployment
// was never created as user hit Ctrl+C before odo could create one.
func (o *DevClient) Cleanup(devfileObj parser.DevfileObj) error {
	var err error
	err = o.kubernetesClient.DeleteDeployment(componentlabels.GetLabels(devfileObj.GetMetadataName(), "app", false))
	if err != nil {
		return err
	}
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
