package docker

import (
	"io"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/openshift/odo/pkg/machineoutput"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/docker/component"
	"github.com/openshift/odo/pkg/lclient"
	"github.com/pkg/errors"
)

// Adapter maps Devfiles to Docker resources and actions
type Adapter struct {
	componentAdapter common.ComponentAdapter
}

// New instantiates a Docker adapter
func New(adapterContext common.AdapterContext, client lclient.Client) Adapter {

	compAdapter := component.New(adapterContext, client)

	return Adapter{
		componentAdapter: compAdapter,
	}
}

// Push creates Docker resources that correspond to the devfile if they don't already exist
func (d Adapter) Push(parameters common.PushParameters) error {

	err := d.componentAdapter.Push(parameters)
	if err != nil {
		return errors.Wrap(err, "Failed to create the component")
	}

	return nil
}

func (a Adapter) CheckSupervisordCtlStatus(command devfilev1.Command) error {
	return nil
}

// DoesComponentExist returns true if a component with the specified name exists
func (d Adapter) DoesComponentExist(cmpName, appName string) (bool, error) {
	return d.componentAdapter.DoesComponentExist(cmpName, appName)
}

// Delete attempts to delete the component with the specified labels, returning an error if it fails
func (d Adapter) Delete(labels map[string]string, show bool, wait bool) error {
	return d.componentAdapter.Delete(labels, show, wait)
}

// Test runs devfile test command
func (d Adapter) Test(testCmd string, show bool) error {
	return d.componentAdapter.Test(testCmd, show)
}

// Log shows logs from component
func (d Adapter) Log(follow bool, command devfilev1.Command) (io.ReadCloser, error) {
	return d.componentAdapter.Log(follow, command)
}

// Exec executes a command in the component
func (d Adapter) Exec(command []string) error {
	return d.componentAdapter.Exec(command)
}

func (d Adapter) ExecCMDInContainer(info common.ComponentInfo, cmd []string, stdOut io.Writer, stdErr io.Writer, stdIn io.Reader, show bool) error {
	return d.componentAdapter.ExecCMDInContainer(info, cmd, stdOut, stdErr, stdIn, show)
}
func (d Adapter) Logger() machineoutput.MachineEventLoggingClient {
	return d.componentAdapter.Logger()
}

func (d Adapter) ComponentInfo(command devfilev1.Command) (common.ComponentInfo, error) {
	return d.componentAdapter.ComponentInfo(command)
}

func (d Adapter) SupervisorComponentInfo(command devfilev1.Command) (common.ComponentInfo, error) {
	return d.componentAdapter.SupervisorComponentInfo(command)
}

// StartContainerStatusWatch outputs container Docker status changes to the console, as used by status command
func (d Adapter) StartContainerStatusWatch() {
	d.componentAdapter.StartContainerStatusWatch()
}

// StartSupervisordCtlStatusWatch outputs supervisord program status changes to the console, as used by status command
func (d Adapter) StartSupervisordCtlStatusWatch() {
	d.componentAdapter.StartSupervisordCtlStatusWatch()
}
