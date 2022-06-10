package remotecmd

import (
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"

	"github.com/redhat-developer/odo/pkg/kclient"
)

// RemoteProcessStatus is an enum type for representing process statuses.
type RemoteProcessStatus int

const (
	// Unknown represents a process for which the status cannot be determined reliably or is not handled yet by us.
	Unknown RemoteProcessStatus = iota + 1

	// Stopped represents a process stopped.
	Stopped

	// Running represents a running process.
	Running
)

// RemoteProcessInfo represents a given remote process linked to a given Devfile command
type RemoteProcessInfo struct {
	// Pid of the process
	Pid int

	// Status of the process
	Status RemoteProcessStatus
}

// CommandOutputHandler is a function that is expected to handle the output and error returned by a command executed.
// It is currently used in StartProcessForCommand and StopProcessForCommand.
type CommandOutputHandler func(stdout []string, stderr []string, err error)

// RemoteProcessHandler is an interface for managing processes that are intended to be executed remotely,
// in Kubernetes/OpenShift containers
type RemoteProcessHandler interface {

	// GetProcessInfoForCommand returns information about the process representing the given Devfile command.
	GetProcessInfoForCommand(
		devfileCmd devfilev1.Command,
		kclient kclient.ClientInterface,
		podName string,
		containerName string,
	) (RemoteProcessInfo, error)

	// StartProcessForCommand starts a process with the provided Devfile command to execute remotely.
	StartProcessForCommand(
		devfileCmd devfilev1.Command,
		kclient kclient.ClientInterface,
		podName string,
		containerName string,
		outputHandler CommandOutputHandler,
	) error

	// StopProcessForCommand stops the process representing the given Devfile command.
	StopProcessForCommand(
		devfileCmd devfilev1.Command,
		kclient kclient.ClientInterface,
		podName string,
		containerName string,
		outputHandler CommandOutputHandler,
	) error
}
