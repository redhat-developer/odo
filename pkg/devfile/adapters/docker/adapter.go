package docker

import (
	"io"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/docker/component"
	"github.com/openshift/odo/pkg/lclient"
	"github.com/pkg/errors"
)

// Adapter maps Devfiles to Docker resources and actions
type Adapter struct {
	componentAdapter common.ComponentAdapter
}

// New instantiates a Docker adapter
func New(adapterContext common.AdapterContext, client lclient.Client) Adapter {

	compAdapter := component.New(adapterContext, client)

	return Adapter{
		componentAdapter: compAdapter,
	}
}

// Push creates Docker resources that correspond to the devfile if they don't already exist
func (d Adapter) Push(parameters common.PushParameters) error {

	err := d.componentAdapter.Push(parameters)
	if err != nil {
		return errors.Wrap(err, "Failed to create the component")
	}

	return nil
}

// DoesComponentExist returns true if a component with the specified name exists
func (d Adapter) DoesComponentExist(cmpName string) (bool, error) {
	return d.componentAdapter.DoesComponentExist(cmpName)
}

// Delete attempts to delete the component with the specified labels, returning an error if it fails
func (d Adapter) Delete(labels map[string]string) error {
	return d.componentAdapter.Delete(labels)
}

// Log show logs from component
func (d Adapter) Log(follow, debug bool) (io.ReadCloser, error) {
	return d.componentAdapter.Log(follow, debug)
}

// Exec executes a command in the component
func (d Adapter) Exec(command []string) error {
	return d.componentAdapter.Exec(command)
}
