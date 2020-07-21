package common

import (
	"fmt"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/pkg/errors"
	"io"
	"k8s.io/klog"
	"strings"
)

type ComponentInfoFactory func(command common.DevfileCommand) (ComponentInfo, error)

type GenericAdapter struct {
	AdapterContext
	client                   ExecClient
	logger                   machineoutput.MachineEventLoggingClient
	componentInfo            ComponentInfoFactory
	supervisordComponentInfo ComponentInfoFactory
}

func NewGenericAdapter(client ExecClient, logger machineoutput.MachineEventLoggingClient, context AdapterContext, ciFactory ComponentInfoFactory, supervisorFactory ComponentInfoFactory) GenericAdapter {
	return GenericAdapter{
		AdapterContext:           context,
		client:                   client,
		logger:                   logger,
		componentInfo:            ciFactory,
		supervisordComponentInfo: supervisorFactory,
	}
}

func (a GenericAdapter) ExecuteCommand(compInfo ComponentInfo, command []string, show bool, consoleOutputStdout *io.PipeWriter, consoleOutputStderr *io.PipeWriter) (err error) {
	return ExecuteCommand(a.client, compInfo, command, show, consoleOutputStdout, consoleOutputStderr)
}

// ExecuteDevfileCommandSynchronously executes the devfile init, build and test command actions synchronously
func (a GenericAdapter) ExecuteDevfileCommandSynchronously(command common.DevfileCommand, show bool) error {
	exe := command.Exec
	var setEnvVariable, cmdLine string
	for _, envVar := range exe.Env {
		setEnvVariable = setEnvVariable + fmt.Sprintf("%v=\"%v\" ", envVar.Name, envVar.Value)
	}
	if setEnvVariable == "" {
		cmdLine = exe.CommandLine
	} else {
		cmdLine = setEnvVariable + "&& " + exe.CommandLine
	}
	// Change to the workdir and execute the command
	var cmd executable
	if exe.WorkingDir != "" {
		// since we are using /bin/sh -c, the command needs to be within a single double quote instance, for example "cd /tmp && pwd"
		cmd = executable{ShellExecutable, "-c", "cd " + exe.WorkingDir + " && " + cmdLine}
	} else {
		cmd = executable{ShellExecutable, "-c", cmdLine}
	}
	return a.Execute(command, show, []executable{cmd})
}

// DefaultCommands returns the devfile commands to execute based on the specified command options
func DefaultCommands(debug, restart bool) []executable {
	cmd := string(DefaultDevfileRunCommand)
	if debug {
		cmd = string(DefaultDevfileDebugCommand)
	}
	klog.V(4).Infof("restart:false, not restarting %s", cmd)

	// with restart false, executing only supervisord start command, if the command is already running, supvervisord will not restart it.
	// if the command is failed or not running supervisord would start it.
	execs := []executable{
		{SupervisordBinaryPath, SupervisordCtlSubCommand, "start", cmd},
	}

	if restart {
		// first stop any running command
		stopSupervisorExec := executable{SupervisordBinaryPath, SupervisordCtlSubCommand, "stop", "all"}
		execs = append([]executable{stopSupervisorExec}, execs...)
	}
	return execs
}

type executable []string

// Execute executes the specified devfile exec command appropriately wrapped into the appropriate sequence of executable, usually by calling DefaultCommands
func (a GenericAdapter) Execute(command common.DevfileCommand, show bool, subcommands []executable) error {
	exe := command.Exec
	msg := fmt.Sprintf("Executing %s command %q, if not running", exe.Id, exe.CommandLine)
	var s *log.Status
	if show {
		s = log.SpinnerNoSpin(msg)
	} else {
		s = log.Spinnerf(msg)
	}
	defer s.End(false)

	for _, subcommand := range subcommands {

		// Emit DevFileCommandExecutionBegin JSON event (if machine output logging is enabled)
		logger := a.logger
		logger.DevFileCommandExecutionBegin(exe.Id, exe.Component, exe.CommandLine, convertGroupKindToString(*exe), machineoutput.TimestampNow())

		// Capture container text and log to the screen as JSON events (machine output only)
		stdoutWriter, stdoutChannel, stderrWriter, stderrChannel := logger.CreateContainerOutputWriter()

		info, err2 := a.componentInfo(command)
		if err2 != nil {
			return errors.Wrapf(err2, "unable to execute the run command")
		}

		err := ExecuteCommand(a.client, info, subcommand, show, stdoutWriter, stderrWriter)

		// Close the writers and wait for an acknowledgement that the reader loop has exited (to ensure we get ALL container output)
		closeWriterAndWaitForAck(stdoutWriter, stdoutChannel, stderrWriter, stderrChannel)

		// Emit close event
		logger.DevFileCommandExecutionComplete(exe.Id, exe.Component, exe.CommandLine, convertGroupKindToString(*exe), machineoutput.TimestampNow(), err)

		if err != nil {
			return errors.Wrapf(err, "unable to execute the run command")
		}
	}

	s.End(true)

	return nil
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

func convertGroupKindToString(exec common.Exec) string {
	if exec.Group == nil {
		return ""
	}
	return string(exec.Group.Kind)
}

// ExecDevFile executes all the commands from the devfile in order: init and build - which are both optional, and a compulsory run.
// Init only runs once when the component is created.
func (a GenericAdapter) ExecDevfile(commandsMap PushCommandsMap, componentExists bool, params PushParameters) (err error) {

	// If nothing has been passed, then the devfile is missing the required run command
	if len(commandsMap) == 0 {
		return errors.New(fmt.Sprint("error executing devfile commands - there should be at least 1 command"))
	}

	// Only add runinit to the expected commands if the component doesn't already exist
	// This would be the case when first running the container
	if !componentExists {
		// Get Init Command
		command, ok := commandsMap[common.InitCommandGroupType]
		if ok {
			err = a.ExecuteDevfileCommandSynchronously(command, params.Show)
			if err != nil {
				return err
			}
		}
	}

	// Get Build Command
	command, ok := commandsMap[common.BuildCommandGroupType]
	if ok {
		err = a.ExecuteDevfileCommandSynchronously(command, params.Show)
		if err != nil {
			return err
		}
	}

	// Get Run or Debug Command
	if params.Debug {
		command, ok = commandsMap[common.DebugCommandGroupType]
	} else {
		command, ok = commandsMap[common.RunCommandGroupType]
	}
	if ok {
		klog.V(4).Infof("Executing devfile command %v", command.Exec.Id)

		// Check if the devfile run component containers have supervisord as the entrypoint.
		// Start the supervisord if the odo component does not exist
		if !componentExists {
			info, err := a.supervisordComponentInfo(command)
			if err != nil {
				a.logger.ReportError(err, machineoutput.TimestampNow())
				return err
			}
			err = ExecuteCommand(a.client, info, []string{SupervisordBinaryPath, "-c", SupervisordConfFile, "-d"}, true, nil, nil)
			if err != nil {
				a.logger.ReportError(err, machineoutput.TimestampNow())
				return err
			}
		}

		return a.Execute(command, params.Show, DefaultCommands(params.Debug, IsRestartRequired(command)))
	}

	return
}

// TODO: Support Composite
// ExecDevfileEvent receives a Devfile Event (PostStart, PreStop etc.) and loops through them
// Each Devfile Command associated with the given event is retrieved, and executed in the container specified
// in the command
func (a GenericAdapter) ExecDevfileEvent(events []string) error {
	if len(events) > 0 {
		commandMap := GetCommandMap(a.Devfile.Data)
		for _, commandName := range events {
			// Convert commandName to lower because GetCommands converts Command.Exec.Id's to lower
			command, ok := commandMap[strings.ToLower(commandName)]
			if !ok {
				return errors.New("unable to find devfile command " + commandName)
			}

			// If composite would go here & recursive loop

			// Execute command in container
			err := a.ExecuteDevfileCommandSynchronously(command, false)
			if err != nil {
				return errors.Wrapf(err, "unable to execute devfile command %s", commandName)
			}
		}
	}
	return nil
}
