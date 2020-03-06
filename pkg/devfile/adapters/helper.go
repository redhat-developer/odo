package adapters

import (
	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes"
	"github.com/openshift/odo/pkg/kclient"
)

// NewPlatformAdapter returns a Devfile adapter for the targeted platform
func NewPlatformAdapter(componentName string, devObj devfile.DevfileObj) (PlatformAdapter, error) {

	adapterContext := common.AdapterContext{
		ComponentName: componentName,
		Devfile:       devObj,
	}

	// Only the kubernetes adapter is implemented at the moment
	// When there are others this function should be updated to retrieve the correct adapter for the desired platform target
	return createKubernetesAdapter(adapterContext, "")
}

// NewPlatformAdapter returns a Devfile adapter for the targeted platform
func NewPlatformAdapterWithNamespace(componentName string, devObj devfile.DevfileObj, namespace string) (PlatformAdapter, error) {

	adapterContext := common.AdapterContext{
		ComponentName: componentName,
		Devfile:       devObj,
	}

	// Only the kubernetes adapter is implemented at the moment
	// When there are others this function should be updated to retrieve the correct adapter for the desired platform target
	return createKubernetesAdapter(adapterContext, namespace)
}

func createKubernetesAdapter(adapterContext common.AdapterContext, namespace string) (PlatformAdapter, error) {
	client, err := kclient.New()
	if err != nil {
		return nil, err
	}

	// If a namespace was passed in
	if namespace != "" {
		client.Namespace = namespace
	}
	return newKubernetesAdapter(adapterContext, *client)
}

func newKubernetesAdapter(adapterContext common.AdapterContext, client kclient.Client) (PlatformAdapter, error) {
	// Feed the common metadata to the platform-specific adapter
	kubernetesAdapter := kubernetes.New(adapterContext, client)

	return kubernetesAdapter, nil
}
