package adapters

import (
	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes"
	"github.com/openshift/odo/pkg/kclient"
)

// NewPlatformAdapter returns a Devfile adapter for the targeted platform
func NewPlatformAdapter(devfile devfile.DevfileObj, odoComponent common.OdoComponent) (PlatformAdapter, error) {
	return newKubernetesAdapter(devfile, odoComponent)
}

func newKubernetesAdapter(devfile devfile.DevfileObj, odoComponent common.OdoComponent) (PlatformAdapter, error) {
	commonMetadata := load(devfile, odoComponent)

	client, err := kclient.New()
	if err != nil {
		return nil, err
	}

	// Feed the common metadata to the platform-specific adapter
	kubernetesAdapter := kubernetes.New(commonMetadata, *client)

	return kubernetesAdapter, nil
}

// load combines metadata that is required for all adapters
func load(devfile devfile.DevfileObj, devFileComponent common.OdoComponent) common.AdapterMetadata {
	return common.AdapterMetadata{
		Devfile: devfile,
		OdoComp: devFileComponent,
	}
}
