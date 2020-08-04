package common

import (
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/machineoutput"
)

// commandExecutor defines the interface adapters must implement to be able to execute commands in a generic way
type commandExecutor interface {
	ExecClient
	// Logger returns the MachineEventLoggingClient associated with this executor
	Logger() machineoutput.MachineEventLoggingClient
	// ComponentInfo retrieves the component information associated with the specified command
	ComponentInfo(command common.DevfileCommand) (ComponentInfo, error)
	// ComponentInfo retrieves the component information associated with the specified command for supervisor initialization purposes
	SupervisorComponentInfo(command common.DevfileCommand) (ComponentInfo, error)
}
