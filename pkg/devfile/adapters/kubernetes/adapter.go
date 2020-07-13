package kubernetes

import (
	"io"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes/component"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/pkg/errors"
)

// Adapter maps Devfiles to Kubernetes resources and actions
type Adapter struct {
	componentAdapter common.ComponentAdapter
}

type KubernetesContext struct {
	Namespace string
}

// New instantiates a kubernetes adapter
func New(adapterContext common.AdapterContext, client kclient.Client) Adapter {

	compAdapter := component.New(adapterContext, client)

	return Adapter{
		componentAdapter: compAdapter,
	}
}

// Push creates Kubernetes resources that correspond to the devfile if they don't already exist
func (k Adapter) Push(parameters common.PushParameters) error {

	err := k.componentAdapter.Push(parameters)
	if err != nil {
		return errors.Wrap(err, "Failed to create the component")
	}

	return nil
}

// DoesComponentExist returns true if a component with the specified name exists
func (k Adapter) DoesComponentExist(cmpName string) (bool, error) {
	return k.componentAdapter.DoesComponentExist(cmpName)
}

// Delete deletes the Kubernetes resources that correspond to the devfile
func (k Adapter) Delete(labels map[string]string) error {

	err := k.componentAdapter.Delete(labels)
	if err != nil {
		return err
	}

	return nil
}

// Log shows log from component
func (k Adapter) Log(follow, debug bool) (io.ReadCloser, error) {
	return k.componentAdapter.Log(follow, debug)
}
