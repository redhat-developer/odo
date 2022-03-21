package kubernetes

import (
	"fmt"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/preference"

	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/component"
)

// Adapter maps Devfiles to Kubernetes resources and actions
type Adapter struct {
	componentAdapter common.ComponentAdapter
}

type KubernetesContext struct {
	Namespace string
}

// New instantiates a kubernetes adapter
func New(adapterContext common.AdapterContext, client kclient.ClientInterface, prefClient preference.Client) Adapter {

	compAdapter := component.New(adapterContext, client, prefClient)

	return Adapter{
		componentAdapter: &compAdapter,
	}
}

// Push creates Kubernetes resources that correspond to the devfile if they don't already exist
func (k Adapter) Push(parameters common.PushParameters) error {

	err := k.componentAdapter.Push(parameters)
	if err != nil {
		return fmt.Errorf("Failed to create the component: %w", err)
	}

	return nil
}

// CheckSupervisordCommandStatus calls the component adapter's CheckSupervisordCommandStatus
func (k Adapter) CheckSupervisordCommandStatus(command devfilev1.Command) error {
	err := k.componentAdapter.CheckSupervisordCommandStatus(command)
	if err != nil {
		return fmt.Errorf("Failed to check the status: %w", err)
	}

	return nil
}
