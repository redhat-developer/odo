package dev

import (
	"fmt"
	"io"

	"github.com/redhat-developer/odo/pkg/dev"
	ododevfile "github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	kcomponent "github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/component"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/watch"
)

type Handler struct {
	clientset   clientset.Clientset
	randomPorts bool
	errOut      io.Writer
}

var _ dev.Handler = (*Handler)(nil)

// RegenerateAdapterAndPush regenerates the adapter and pushes the files to remote pod
func (o *Handler) RegenerateAdapterAndPush(pushParams adapters.PushParameters, watchParams watch.WatchParameters, componentStatus *watch.ComponentStatus) error {
	var adapter kcomponent.ComponentAdapter

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

func (o *Handler) regenerateComponentAdapterFromWatchParams(parameters watch.WatchParameters) (kcomponent.ComponentAdapter, error) {
	devObj, err := ododevfile.ParseAndValidateFromFileWithVariables(location.DevfileLocation(""), parameters.Variables)
	if err != nil {
		return nil, err
	}

	return kcomponent.NewKubernetesAdapter(
		o.clientset.KubernetesClient,
		o.clientset.PreferenceClient,
		o.clientset.PortForwardClient,
		o.clientset.BindingClient,
		kcomponent.AdapterContext{
			ComponentName: parameters.ComponentName,
			Context:       parameters.Path,
			AppName:       parameters.ApplicationName,
			Devfile:       devObj,
			FS:            o.clientset.FS,
		},
		o.clientset.KubernetesClient.GetCurrentNamespace(),
	), nil
}
