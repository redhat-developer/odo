package remotecmd

// RemoteProcessStatus is an enum type for representing process statuses.
type RemoteProcessStatus string

const (
	// Unknown represents a process for which the status cannot be determined reliably or is not handled yet by us.
	Unknown RemoteProcessStatus = "unknown"

	// Starting represents a process that is just about to start.
	Starting = "starting"

	// Stopped represents a process stopped.
	Stopped = "stopped"

	// Errored represents a process that errored out, i.e. exited with a non-zero status code.
	Errored = "errored"

	// Running represents a running process.
	Running = "running"
)

const (
	// ShellExecutable is the shell executable
	ShellExecutable = "/bin/sh"
)

// RemoteProcessInfo represents a given remote process linked to a given Devfile command
type RemoteProcessInfo struct {
	// Pid of the process
	Pid int

	// Status of the process
	Status RemoteProcessStatus
}

// CommandDefinition represents the structure of any given command that would be handled by implementations of RemoteProcessHandler.
type CommandDefinition struct {
	// Id is any unique (and short) identifier that helps manage the process associated to this command.
	Id string

	// PidDirectory is the directory where the PID file for this process will be stored.
	// The directory needs to be present in the remote container and be writable by the user (in the container) executing the command.
	PidDirectory string

	// WorkingDir is the working directory from which the command should get executed.
	WorkingDir string

	// EnvVars are environment variables to set.
	EnvVars []CommandEnvVar

	// CmdLine is the command-line that will get executed.
	CmdLine string
}

// CommandEnvVar represents an environment variable used as part of running any CommandDefinition.
type CommandEnvVar struct {

	// Key of the environment variable.
	Key string

	// Value of the environment variable.
	Value string
}

// CommandOutputHandler is a function that is expected to handle the output and error returned by a command executed.
type CommandOutputHandler func(status RemoteProcessStatus, stdout []string, stderr []string, err error)
