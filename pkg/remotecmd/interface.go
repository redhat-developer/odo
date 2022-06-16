package remotecmd

import (
	"github.com/redhat-developer/odo/pkg/kclient"
)

// RemoteProcessHandler is an interface for managing processes that are intended to be executed remotely,
// in Kubernetes/OpenShift containers
type RemoteProcessHandler interface {

	// GetProcessInfoForCommand returns information about the process representing the given command.
	GetProcessInfoForCommand(
		def CommandDefinition,
		kclient kclient.ClientInterface,
		podName string,
		containerName string,
	) (RemoteProcessInfo, error)

	// StartProcessForCommand starts a process with the provided Devfile command to execute remotely.
	StartProcessForCommand(
		def CommandDefinition,
		kclient kclient.ClientInterface,
		podName string,
		containerName string,
		outputHandler CommandOutputHandler,
	) error

	// StopProcessForCommand stops the process representing the given Devfile command.
	StopProcessForCommand(
		def CommandDefinition,
		kclient kclient.ClientInterface,
		podName string,
		containerName string,
	) error
}
