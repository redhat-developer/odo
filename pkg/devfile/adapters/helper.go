package adapters

import (
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes"
	"github.com/openshift/odo/pkg/kclient"
)

// NewPlatformAdapter returns a Devfile adapter for the targeted platform
func NewPlatformAdapter(adapterMetadata common.AdapterMetadata) (PlatformAdapter, error) {
	// Only the kubernetes adapter is implemented at the moment
	// When there are others this function should be updated to retrieve the correct adapter for the desired platform target
	return newKubernetesAdapter(adapterMetadata)
}

func newKubernetesAdapter(adapterMetadata common.AdapterMetadata) (PlatformAdapter, error) {
	client, err := kclient.New()
	if err != nil {
		return nil, err
	}

	// Feed the common metadata to the platform-specific adapter
	kubernetesAdapter := kubernetes.New(adapterMetadata, *client)

	return kubernetesAdapter, nil
}
