package common

import (
	"io"
	"strings"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

// ComponentInfoFactory defines a type for a function which creates a ComponentInfo based on the information provided by the specified DevfileCommand.
// This is used by adapters to provide the proper ComponentInfo identifying which component (including supervisor) to target when executing the command.
type ComponentInfoFactory func(command devfilev1.Command) (ComponentInfo, error)

// GenericAdapter provides common code that can be reused by adapters allowing them to focus on more specific behavior
type GenericAdapter struct {
	AdapterContext
	client                   ExecClient
	logger                   machineoutput.MachineEventLoggingClient
	componentInfo            ComponentInfoFactory
	supervisordComponentInfo ComponentInfoFactory
}

// NewGenericAdapter creates a new GenericAdapter instance based on the provided parameters. Client code must call InitWith on
// the newly created instance to finish the setup, providing the child implementation as parameter
func NewGenericAdapter(client ExecClient, context AdapterContext) *GenericAdapter {
	return &GenericAdapter{
		AdapterContext: context,
		client:         client,
		logger:         machineoutput.NewMachineEventLoggingClient(),
	}
}

// InitWith finishes the GenericAdapter setup after the rest of the adapter is created. This must be called before the adapter
// implementation is used and the specific implementation must be passed as parameter.
func (a *GenericAdapter) InitWith(executor commandExecutor) {
	a.componentInfo = executor.ComponentInfo
	a.supervisordComponentInfo = executor.SupervisorComponentInfo
}

func (a GenericAdapter) ExecCMDInContainer(info ComponentInfo, cmd []string, stdOut io.Writer, stdErr io.Writer, stdIn io.Reader, show bool) error {
	return a.client.ExecCMDInContainer(info, cmd, stdOut, stdErr, stdIn, show)
}

func (a GenericAdapter) Logger() machineoutput.MachineEventLoggingClient {
	return a.logger
}

func (a *GenericAdapter) SetLogger(loggingClient machineoutput.MachineEventLoggingClient) {
	a.logger = loggingClient
}

func (a GenericAdapter) ComponentInfo(command devfilev1.Command) (ComponentInfo, error) {
	return a.componentInfo(command)
}

func (a GenericAdapter) SupervisorComponentInfo(command devfilev1.Command) (ComponentInfo, error) {
	return a.supervisordComponentInfo(command)
}

// ExecuteCommand simply calls exec.ExecuteCommand using the GenericAdapter's client
func (a GenericAdapter) ExecuteCommand(compInfo ComponentInfo, command []string, show bool, consoleOutputStdout *io.PipeWriter, consoleOutputStderr *io.PipeWriter) (err error) {
	return ExecuteCommand(a.client, compInfo, command, show, consoleOutputStdout, consoleOutputStderr)
}

// ExecuteDevfileCommand executes the devfile init, build and test command actions synchronously
func (a GenericAdapter) ExecuteDevfileCommand(command devfilev1.Command, show bool) error {
	commands, err := a.Devfile.Data.GetCommands(common.DevfileOptions{})
	if err != nil {
		return err
	}

	c, err := New(command, GetCommandsMap(commands), a)
	if err != nil {
		return err
	}
	return c.Execute(show)
}

// closeWriterAndWaitForAck closes the PipeWriter and then waits for a channel response from the ContainerOutputWriter (indicating that the reader had closed).
// This ensures that we always get the full stderr/stdout output from the container process BEFORE we output the devfileCommandExecution event.
func closeWriterAndWaitForAck(stdoutWriter *io.PipeWriter, stdoutChannel chan interface{}, stderrWriter *io.PipeWriter, stderrChannel chan interface{}) {
	if stdoutWriter != nil {
		_ = stdoutWriter.Close()
		<-stdoutChannel
	}
	if stderrWriter != nil {
		_ = stderrWriter.Close()
		<-stderrChannel
	}
}

func convertGroupKindToString(exec *devfilev1.ExecCommand) string {
	if exec.Group == nil {
		return ""
	}
	return string(exec.Group.Kind)
}

// ExecDevFile executes all the commands from the devfile in order: init and build - which are both optional, and a compulsory run.
// Init only runs once when the component is created.
func (a GenericAdapter) ExecDevfile(commandsMap PushCommandsMap, componentExists bool, params PushParameters) (err error) {
	// Need to get mapping of all commands in the devfile since the composite command may reference any exec or composite command in the devfile
	devfileCommands, err := a.Devfile.Data.GetCommands(common.DevfileOptions{})
	if err != nil {
		return err
	}

	devfileCommandMap := GetCommandsMap(devfileCommands)

	// If nothing has been passed, then the devfile is missing the required run command
	if len(commandsMap) == 0 {
		return errors.New("error executing devfile commands - there should be at least 1 command")
	}

	commands := make([]command, 0, 7)

	// Get Build Command
	commands, err = a.addToComposite(commandsMap, devfilev1.BuildCommandGroupKind, devfileCommandMap, commands)
	if err != nil {
		return err
	}

	group := devfilev1.RunCommandGroupKind
	defaultCmd := string(DefaultDevfileRunCommand)

	if params.Debug {
		group = devfilev1.DebugCommandGroupKind
		defaultCmd = string(DefaultDevfileDebugCommand)

	}

	if command, ok := commandsMap[group]; ok {
		// if the component doesn't exist, initialize the supervisor if needed
		if !componentExists {
			if cmd, err := newSupervisorInitCommand(command, a); cmd != nil {
				if err != nil {
					return err
				}
				commands = append(commands, cmd)
			}
		}

		restart := IsRestartRequired(util.SafeGetBool(command.Exec.HotReloadCapable), params.RunModeChanged)

		// if we need to restart, issue supervisor command to stop all running commands first
		// we do not need to restart Hot reload capable commands
		if componentExists {
			if restart {
				klog.V(2).Infof("supervisord stop command to restart or start other command")
				if cmd, err := newSupervisorStopCommand(command, a); cmd != nil {
					if err != nil {
						return err
					}
					commands = append(commands, cmd)
				}
			} else {
				klog.V(2).Infof("command is hot reload capable, not restarting %s", defaultCmd)
			}
		}

		// with restart false, executing only supervisord start command, if the command is already running, supvervisord will not restart it.
		// if the command is failed or not running supervisord would start it.
		if cmd, err := newSupervisorStartCommand(command, defaultCmd, a, restart); cmd != nil {
			if err != nil {
				return err
			}
			commands = append(commands, cmd)
		}

		c := newCompositeCommand(commands...)
		return c.Execute(params.Show)
	}

	return nil
}

func (a GenericAdapter) addToComposite(commandsMap PushCommandsMap, groupType devfilev1.CommandGroupKind, devfileCommandMap map[string]devfilev1.Command, commands []command) ([]command, error) {
	command, ok := commandsMap[groupType]
	if ok {
		if c, err := New(command, devfileCommandMap, a); err == nil {
			commands = append(commands, c)
		} else {
			return commands, err
		}
	}
	return commands, nil
}

// ExecDevfileEvent receives a Devfile Event (PostStart, PreStop etc.) and loops through them
// Each Devfile Command associated with the given event is retrieved, and executed in the container specified
// in the command
func (a GenericAdapter) ExecDevfileEvent(events []string, eventType DevfileEventType, show bool) error {
	if len(events) > 0 {
		log.Infof("\nExecuting %s event commands for component %s", string(eventType), a.ComponentName)
		commands, err := a.Devfile.Data.GetCommands(common.DevfileOptions{})
		if err != nil {
			return err
		}

		commandMap := GetCommandsMap(commands)
		for _, commandName := range events {
			// Convert commandName to lower because GetCommands converts Command.Exec.Id's to lower
			command, ok := commandMap[strings.ToLower(commandName)]
			if !ok {
				return errors.New("unable to find devfile command " + commandName)
			}

			c, err := New(command, commandMap, a)
			if err != nil {
				return err
			}
			// Execute command in container
			err = c.Execute(show)
			if err != nil {
				return errors.Wrapf(err, "unable to execute devfile command %s", commandName)
			}
		}
	}
	return nil
}

func (a GenericAdapter) ApplyComponent(component string) error {
	return nil
}
