package common

import (
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/machineoutput"
)

// commandExecutor defines the interface adapters must implement to be able to execute commands in a generic way
type commandExecutor interface {
	ExecClient
	Logger() machineoutput.MachineEventLoggingClient
	ComponentInfo(command common.DevfileCommand) (ComponentInfo, error)
	SupervisorComponentInfo(command common.DevfileCommand) (ComponentInfo, error)
}
