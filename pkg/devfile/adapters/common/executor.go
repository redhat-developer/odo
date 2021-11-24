package common

import (
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/openshift/odo/pkg/machineoutput"
)

// commandExecutor defines the interface adapters must implement to be able to execute commands in a generic way
type commandExecutor interface {
	ExecClient
	ApplyClient
	// Logger returns the MachineEventLoggingClient associated with this executor
	Logger() machineoutput.MachineEventLoggingClient
	// ComponentInfo retrieves the component information associated with the specified command
	ComponentInfo(command devfilev1.Command) (ComponentInfo, error)
	// ComponentInfo retrieves the component information associated with the specified command for supervisor initialization purposes
	SupervisorComponentInfo(command devfilev1.Command) (ComponentInfo, error)
}
