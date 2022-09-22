package dev

import (
	"context"
	"fmt"
	"io"

	"github.com/redhat-developer/odo/pkg/binding"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/portForward"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"

	"github.com/devfile/library/pkg/devfile/parser"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/component"
	"github.com/redhat-developer/odo/pkg/devfile/location"
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
	out io.Writer,
	errOut io.Writer,
	options StartOptions,
) error {
	klog.V(4).Infoln("Creating new adapter")

	adapter := component.NewKubernetesAdapter(
		o.kubernetesClient, o.prefClient, o.portForwardClient, o.bindingClient,
		component.AdapterContext{
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
		IgnoredFiles:    options.IgnorePaths,
		Debug:           options.Debug,
		DevfileBuildCmd: options.BuildCommand,
		DevfileRunCmd:   options.RunCommand,
		RandomPorts:     options.RandomPorts,
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
		DevfileWatchHandler: o.regenerateAdapterAndPush,
		EnvSpecificInfo:     envSpecificInfo,
		FileIgnores:         options.IgnorePaths,
		InitialDevfileObj:   devfileObj,
		Debug:               options.Debug,
		DevfileBuildCmd:     options.BuildCommand,
		DevfileRunCmd:       options.RunCommand,
		Variables:           options.Variables,
		RandomPorts:         options.RandomPorts,
		WatchFiles:          options.WatchFiles,
		ErrOut:              errOut,
	}

	return o.watchClient.WatchAndPush(out, watchParameters, ctx, componentStatus)
}

// RegenerateAdapterAndPush regenerates the adapter and pushes the files to remote pod
func (o *DevClient) regenerateAdapterAndPush(pushParams adapters.PushParameters, watchParams watch.WatchParameters, componentStatus *watch.ComponentStatus) error {
	var adapter component.ComponentAdapter

	adapter, err := o.regenerateComponentAdapterFromWatchParams(watchParams)
	if err != nil {
		return fmt.Errorf("unable to generate component from watch parameters: %w", err)
	}

	err = adapter.Push(pushParams, componentStatus)
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
		component.AdapterContext{
			ComponentName: parameters.ComponentName,
			Context:       parameters.Path,
			AppName:       parameters.ApplicationName,
			Devfile:       devObj,
			FS:            o.filesystem,
		},
	), nil
}
