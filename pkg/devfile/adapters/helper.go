package adapters

import (
	"fmt"

	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/docker"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/lclient"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/util/pushtarget"
)

// NewComponentAdapter returns a Devfile adapter for the targeted platform
func NewComponentAdapter(componentName string, context string, appName string, devObj devfileParser.DevfileObj, platformContext interface{}) (common.ComponentAdapter, error) {

	adapterContext := common.AdapterContext{
		ComponentName: componentName,
		Context:       context,
		AppName:       appName,
		Devfile:       devObj,
	}

	// If the pushtarget is set to Docker, initialize the Docker adapter, otherwise initialize the Kubernetes adapter
	if pushtarget.IsPushTargetDocker() {
		return createDockerAdapter(adapterContext)
	}

	kc, ok := platformContext.(kubernetes.KubernetesContext)
	if !ok {
		return nil, fmt.Errorf("Error retrieving context for Kubernetes")
	}
	return createKubernetesAdapter(adapterContext, kc.Namespace)

}

func createKubernetesAdapter(adapterContext common.AdapterContext, namespace string) (common.ComponentAdapter, error) {
	client, err := occlient.New()
	if err != nil {
		return nil, err
	}

	kClient, err := kclient.New()
	if err != nil {
		return nil, err
	}
	client.SetKubeClient(kClient)

	// If a namespace was passed in
	if namespace != "" {
		client.Namespace = namespace
	}
	return newKubernetesAdapter(adapterContext, *client)
}

func newKubernetesAdapter(adapterContext common.AdapterContext, client occlient.Client) (common.ComponentAdapter, error) {
	// Feed the common metadata to the platform-specific adapter
	kubernetesAdapter := kubernetes.New(adapterContext, client)

	return kubernetesAdapter, nil
}

func createDockerAdapter(adapterContext common.AdapterContext) (common.ComponentAdapter, error) {
	client, err := lclient.New()
	if err != nil {
		return nil, err
	}

	return newDockerAdapter(adapterContext, *client)
}

func newDockerAdapter(adapterContext common.AdapterContext, client lclient.Client) (common.ComponentAdapter, error) {
	dockerAdapter := docker.New(adapterContext, client)
	return dockerAdapter, nil
}
