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
func (d Adapter) Push(path string, ignoredFiles []string, forceBuild bool, globExps []string) error {

	err := d.componentAdapter.Push(path, ignoredFiles, forceBuild, globExps)
	if err != nil {
		return errors.Wrap(err, "Failed to create the component")
	}

	return nil
}
