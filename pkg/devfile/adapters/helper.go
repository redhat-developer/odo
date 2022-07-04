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

	if kc.Namespace != "" {
		kubernetesClient.SetNamespace(kc.Namespace)
	}

	kubernetesAdapter := kubernetes.New(adapterContext, kubernetesClient, prefClient, portForwardClient, randomPorts, errOut)
	return kubernetesAdapter, nil
}
