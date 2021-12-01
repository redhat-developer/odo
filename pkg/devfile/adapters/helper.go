package adapters

import (
	"fmt"

	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/occlient"
)

// NewComponentAdapter returns a Devfile adapter for the targeted platform
func NewComponentAdapter(componentName string, context string, appName string, devObj devfileParser.DevfileObj, platformContext interface{}) (common.ComponentAdapter, error) {

	adapterContext := common.AdapterContext{
		ComponentName: componentName,
		Context:       context,
		AppName:       appName,
		Devfile:       devObj,
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
		kClient.Namespace = namespace
	}
	return newKubernetesAdapter(adapterContext, *client)
}

func newKubernetesAdapter(adapterContext common.AdapterContext, client occlient.Client) (common.ComponentAdapter, error) {
	// Feed the common metadata to the platform-specific adapter
	kubernetesAdapter := kubernetes.New(adapterContext, client)

	return kubernetesAdapter, nil
}
