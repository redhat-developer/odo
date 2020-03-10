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

// Start creates Kubernetes resources that correspond to the devfile if they don't already exist
func (k Adapter) Start(path string, out io.Writer, ignoredFiles []string, forceBuild bool, globExps []string, show bool) error {

	err := k.componentAdapter.Start(path, out, ignoredFiles, forceBuild, globExps, show)
	if err != nil {
		return errors.Wrap(err, "Failed to start the component")
	}

	return nil
}
