package kubernetes

import (
	"io"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/openshift/odo/pkg/machineoutput"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes/component"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/pkg/errors"
)

// Adapter maps Devfiles to Kubernetes resources and actions
type Adapter struct {
	componentAdapter common.ComponentAdapter
}

type KubernetesContext struct {
	Namespace string
}

// New instantiates a kubernetes adapter
func New(adapterContext common.AdapterContext, client occlient.Client) Adapter {

	compAdapter := component.New(adapterContext, client)

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

func (k Adapter) Deploy() error {
	return k.componentAdapter.Deploy()
}

// CheckSupervisordCommandStatus calls the component adapter's CheckSupervisordCommandStatus
func (k Adapter) CheckSupervisordCommandStatus(command devfilev1.Command) error {
	err := k.componentAdapter.CheckSupervisordCommandStatus(command)
	if err != nil {
		return errors.Wrap(err, "Failed to check the status")
	}

	return nil
}

// DoesComponentExist returns true if a component with the specified name exists in the given app
func (k Adapter) DoesComponentExist(cmpName, appName string) (bool, error) {
	return k.componentAdapter.DoesComponentExist(cmpName, appName)
}

// Delete deletes the Kubernetes resources that correspond to the devfile
func (k Adapter) Delete(labels map[string]string, show bool, wait bool) error {

	err := k.componentAdapter.Delete(labels, show, wait)
	if err != nil {
		return err
	}

	return nil
}

// Test runs the devfile test command
func (k Adapter) Test(testCmd string, show bool) error {
	return k.componentAdapter.Test(testCmd, show)
}

// Log shows log from component
func (k Adapter) Log(follow bool, command devfilev1.Command) (io.ReadCloser, error) {
	return k.componentAdapter.Log(follow, command)
}

// Exec executes a command in the component
func (k Adapter) Exec(command []string) error {
	return k.componentAdapter.Exec(command)
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

// StartContainerStatusWatch outputs Kubernetes pod/container status changes to the console, as used by the status command
func (k Adapter) StartContainerStatusWatch() {
	k.componentAdapter.StartContainerStatusWatch()
}

// StartSupervisordCtlStatusWatch outputs supervisord program status changes to the console, as used by the status command
func (k Adapter) StartSupervisordCtlStatusWatch() {
	k.componentAdapter.StartSupervisordCtlStatusWatch()
}

func (k Adapter) ApplyComponent(component string) error {
	return k.componentAdapter.ApplyComponent(component)
}
