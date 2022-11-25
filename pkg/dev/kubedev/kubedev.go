package kubedev

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/redhat-developer/odo/pkg/binding"
	_delete "github.com/redhat-developer/odo/pkg/component/delete"
	"github.com/redhat-developer/odo/pkg/dev"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/portForward"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/sync"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"

	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/component"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/watch"
)

const (
	promptMessage = `
[Ctrl+c] - Exit and delete resources from the cluster
     [p] - Manually apply local changes to the application on the cluster
`
)

type DevClient struct {
	kubernetesClient  kclient.ClientInterface
	prefClient        preference.Client
	portForwardClient portForward.Client
	watchClient       watch.Client
	bindingClient     binding.Client
	syncClient        sync.Client
	filesystem        filesystem.Filesystem
	execClient        exec.Client
	deleteClient      _delete.Client
}

var _ dev.Client = (*DevClient)(nil)

func NewDevClient(
	kubernetesClient kclient.ClientInterface,
	prefClient preference.Client,
	portForwardClient portForward.Client,
	watchClient watch.Client,
	bindingClient binding.Client,
	syncClient sync.Client,
	filesystem filesystem.Filesystem,
	execClient exec.Client,
	deleteClient _delete.Client,
) *DevClient {
	return &DevClient{
		kubernetesClient:  kubernetesClient,
		prefClient:        prefClient,
		portForwardClient: portForwardClient,
		watchClient:       watchClient,
		bindingClient:     bindingClient,
		syncClient:        syncClient,
		filesystem:        filesystem,
		execClient:        execClient,
		deleteClient:      deleteClient,
	}
}

func (o *DevClient) Start(
	ctx context.Context,
	out io.Writer,
	errOut io.Writer,
	options dev.StartOptions,
) error {
	klog.V(4).Infoln("Creating new adapter")

	var (
		devfileObj    = odocontext.GetDevfileObj(ctx)
		devfilePath   = odocontext.GetDevfilePath(ctx)
		path          = filepath.Dir(devfilePath)
		componentName = odocontext.GetComponentName(ctx)
	)

	adapter := component.NewKubernetesAdapter(
		o.kubernetesClient, o.prefClient, o.portForwardClient, o.bindingClient, o.syncClient, o.execClient,
		component.AdapterContext{
			ComponentName: componentName,
			Context:       path,
			AppName:       odocontext.GetApplication(ctx),
			Devfile:       *devfileObj,
			FS:            o.filesystem,
		})

	pushParameters := adapters.PushParameters{
		Path:            path,
		IgnoredFiles:    options.IgnorePaths,
		Debug:           options.Debug,
		DevfileBuildCmd: options.BuildCommand,
		DevfileRunCmd:   options.RunCommand,
		RandomPorts:     options.RandomPorts,
		ErrOut:          errOut,
	}

	klog.V(4).Infoln("Creating inner-loop resources for the component")
	componentStatus := watch.ComponentStatus{}
	err := adapter.Push(ctx, pushParameters, &componentStatus)
	if err != nil {
		return err
	}
	klog.V(4).Infoln("Successfully created inner-loop resources")

	watchParameters := watch.WatchParameters{
		DevfilePath:         devfilePath,
		Path:                path,
		ComponentName:       componentName,
		ApplicationName:     odocontext.GetApplication(ctx),
		DevfileWatchHandler: o.regenerateAdapterAndPush,
		FileIgnores:         options.IgnorePaths,
		InitialDevfileObj:   *devfileObj,
		Debug:               options.Debug,
		DevfileBuildCmd:     options.BuildCommand,
		DevfileRunCmd:       options.RunCommand,
		Variables:           options.Variables,
		RandomPorts:         options.RandomPorts,
		WatchFiles:          options.WatchFiles,
		WatchCluster:        true,
		ErrOut:              errOut,
		PromptMessage:       promptMessage,
	}

	return o.watchClient.WatchAndPush(out, watchParameters, ctx, componentStatus)
}

// RegenerateAdapterAndPush regenerates the adapter and pushes the files to remote pod
func (o *DevClient) regenerateAdapterAndPush(ctx context.Context, pushParams adapters.PushParameters, watchParams watch.WatchParameters, componentStatus *watch.ComponentStatus) error {
	var adapter component.ComponentAdapter

	adapter, err := o.regenerateComponentAdapterFromWatchParams(watchParams)
	if err != nil {
		return fmt.Errorf("unable to generate component from watch parameters: %w", err)
	}

	err = adapter.Push(ctx, pushParams, componentStatus)
	if err != nil {
		return fmt.Errorf("watch command was unable to push component: %w", err)
	}

	return nil
}

func (o *DevClient) regenerateComponentAdapterFromWatchParams(parameters watch.WatchParameters) (component.ComponentAdapter, error) {
	devObj, err := devfile.ParseAndValidateFromFileWithVariables(location.DevfileLocation(""), parameters.Variables)
	if err != nil {
		return nil, err
	}

	return component.NewKubernetesAdapter(
		o.kubernetesClient,
		o.prefClient,
		o.portForwardClient,
		o.bindingClient,
		o.syncClient,
		o.execClient,
		component.AdapterContext{
			ComponentName: parameters.ComponentName,
			Context:       parameters.Path,
			AppName:       parameters.ApplicationName,
			Devfile:       devObj,
			FS:            o.filesystem,
		},
	), nil
}
