package remotecmd

import "context"

// RemoteProcessHandler is an interface for managing processes that are intended to be executed remotely,
// independently of container orchestrator
type RemoteProcessHandler interface {

	// GetProcessInfoForCommand returns information about the process representing the given command.
	GetProcessInfoForCommand(ctx context.Context, def CommandDefinition, podName string, containerName string) (RemoteProcessInfo, error)

	// StartProcessForCommand starts a process with the provided Devfile command to execute remotely.
	StartProcessForCommand(ctx context.Context, def CommandDefinition, podName string, containerName string, outputHandler CommandOutputHandler) error

	// StopProcessForCommand stops the process representing the given Devfile command.
	StopProcessForCommand(ctx context.Context, def CommandDefinition, podName string, containerName string) error
}
