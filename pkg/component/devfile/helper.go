package devfile

import (
	"github.com/openshift/odo/pkg/component/devfile/adapters"
	"github.com/openshift/odo/pkg/component/devfile/adapters/common"
	"github.com/openshift/odo/pkg/component/devfile/adapters/kubernetes"
	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/kclient"
)

// NewPlatformAdapter returns a Devfile adapter for the targeted platform
func NewPlatformAdapter(devfile devfile.DevfileObj, odoComponent common.DevfileComponent) (adapters.PlatformAdapter, error) {
	return newKubernetesAdapter(devfile, odoComponent)
}

func newKubernetesAdapter(devfile devfile.DevfileObj, odoComponent common.DevfileComponent) (adapters.PlatformAdapter, error) {
	commonAdapter, err := load(devfile, odoComponent)

	client, err := kclient.New()
	if err != nil {
		return nil, err
	}

	// Feed the common struct to the platform-specific adapter
	kubernetesAdapter := kubernetes.Adapter{
		DevfileAdapter: commonAdapter,
		Client:         *client,
	}

	return kubernetesAdapter, nil
}

// load takes metadata that is common to all adapters
func load(devfile devfile.DevfileObj, devFileComponent common.DevfileComponent) (common.DevfileAdapter, error) {
	commonAdapter := common.DevfileAdapter{
		Devfile:     devfile,
		DevfileComp: devFileComponent,
	}

	return commonAdapter, nil
}
