package docker

import (
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

func (k Adapter) Build(parameters common.BuildParameters) error {
	return errors.New("Deploy command not supported when building image using pushTarget=Docker")
}

func (k Adapter) Deploy(parameters common.DeployParameters) error {
	return errors.New("Deploy command not supported when using pushTarget=Docker")
}

func (k Adapter) DeployDelete(manifest []byte) error {
	return errors.New("Deploy delete command not supported when using pushTarget=Docker")
}

// DoesComponentExist returns true if a component with the specified name exists
func (d Adapter) DoesComponentExist(cmpName string) bool {
	return d.componentAdapter.DoesComponentExist(cmpName)
}

// Delete attempts to delete the component with the specified labels, returning an error if it fails
func (d Adapter) Delete(labels map[string]string) error {
	return d.componentAdapter.Delete(labels)
}
