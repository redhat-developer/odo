package kubernetes

import (
	"io"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/preference"

	"github.com/pkg/errors"
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
		return errors.Wrap(err, "Failed to create the component")
	}

	return nil
}

// CheckSupervisordCommandStatus calls the component adapter's CheckSupervisordCommandStatus
func (k Adapter) CheckSupervisordCommandStatus(command devfilev1.Command) error {
	err := k.componentAdapter.CheckSupervisordCommandStatus(command)
	if err != nil {
		return errors.Wrap(err, "Failed to check the status")
	}

	return nil
}

func (k Adapter) ExecCMDInContainer(info common.ComponentInfo, cmd []string, stdOut io.Writer, stdErr io.Writer, stdIn io.Reader, show bool) error {
	return k.componentAdapter.ExecCMDInContainer(info, cmd, stdOut, stdErr, stdIn, show)
}
func (k Adapter) Logger() machineoutput.MachineEventLoggingClient {
	return k.componentAdapter.Logger()
}

func (k Adapter) ComponentInfo(command devfilev1.Command) (common.ComponentInfo, error) {
	return k.componentAdapter.ComponentInfo(command)
}

func (k Adapter) SupervisorComponentInfo(command devfilev1.Command) (common.ComponentInfo, error) {
	return k.componentAdapter.SupervisorComponentInfo(command)
}

func (k Adapter) ApplyComponent(component string) error {
	return k.componentAdapter.ApplyComponent(component)
}

func (k Adapter) UnApplyComponent(component string) error {
	return k.componentAdapter.UnApplyComponent(component)
}
