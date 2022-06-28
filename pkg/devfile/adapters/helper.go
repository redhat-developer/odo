package adapters

import (
	"errors"
	"io"

	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/portForward"
	"github.com/redhat-developer/odo/pkg/preference"
)

// NewComponentAdapter returns a Devfile adapter for the targeted platform
func NewComponentAdapter(
	kubernetesClient kclient.ClientInterface,
	prefClient preference.Client,
	portForwardClient portForward.Client,
	componentName string,
	context string,
	appName string,
	devObj devfileParser.DevfileObj,
	platformContext interface{},
	randomPorts bool,
	errOut io.Writer,
) (common.ComponentAdapter, error) {

	adapterContext := common.AdapterContext{
		ComponentName: componentName,
		Context:       context,
		AppName:       appName,
		Devfile:       devObj,
	}

	kc, ok := platformContext.(kubernetes.KubernetesContext)
	if !ok {
		return nil, errors.New("error retrieving context for Kubernetes")
	}
	return createKubernetesAdapter(adapterContext, kubernetesClient, prefClient, portForwardClient, kc.Namespace, randomPorts, errOut)

}

func createKubernetesAdapter(
	adapterContext common.AdapterContext,
	kubernetesClient kclient.ClientInterface,
	prefClient preference.Client,
	portForwardClient portForward.Client,
	namespace string,
	randomPorts bool,
	errOut io.Writer,
) (common.ComponentAdapter, error) {
	if namespace != "" {
		kubernetesClient.SetNamespace(namespace)
	}
	return newKubernetesAdapter(adapterContext, kubernetesClient, prefClient, portForwardClient, randomPorts, errOut)
}

func newKubernetesAdapter(
	adapterContext common.AdapterContext,
	client kclient.ClientInterface,
	prefClient preference.Client,
	portForwardClient portForward.Client,
	randomPorts bool,
	errOut io.Writer,
) (common.ComponentAdapter, error) {
	// Feed the common metadata to the platform-specific adapter
	kubernetesAdapter := kubernetes.New(adapterContext, client, prefClient, portForwardClient, randomPorts, errOut)

	return kubernetesAdapter, nil
}
