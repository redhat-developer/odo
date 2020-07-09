package exec

import (
	"fmt"
	"io"

	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/pkg/errors"
)

// ExecuteDevfileBuildAction executes the devfile build command action
func ExecuteDevfileBuildAction(client ExecClient, exec common.Exec, commandName string, compInfo adaptersCommon.ComponentInfo, show bool, machineEventLogger machineoutput.MachineEventLoggingClient) error {
	var s *log.Status
	var setEnvVariable, command string
	for _, envVar := range exec.Env {
		setEnvVariable = setEnvVariable + fmt.Sprintf("%v=\"%v\" ", envVar.Name, envVar.Value)
	}
	if setEnvVariable == "" {
		command = exec.CommandLine
	} else {
		command = setEnvVariable + " && " + exec.CommandLine
	}
	// Change to the workdir and execute the command
	var cmdArr []string
	if exec.WorkingDir != "" {
		// since we are using /bin/sh -c, the command needs to be within a single double quote instance, for example "cd /tmp && pwd"
		cmdArr = []string{adaptersCommon.ShellExecutable, "-c", "cd " + exec.WorkingDir + " && " + command}
	} else {
		cmdArr = []string{adaptersCommon.ShellExecutable, "-c", command}
	}

	if show {
		s = log.SpinnerNoSpin("Executing " + commandName + " command " + fmt.Sprintf("%q", exec.CommandLine))
	} else {
		s = log.Spinnerf("Executing %s command %q", commandName, exec.CommandLine)
	}

	defer s.End(false)

	// Emit DevFileCommandExecutionBegin JSON event (if machine output logging is enabled)
	machineEventLogger.DevFileCommandExecutionBegin(exec.Id, exec.Component, command, convertGroupKindToString(exec), machineoutput.TimestampNow())

	// Capture container text and log to the screen as JSON events (machine output only)
	stdoutWriter, stdoutChannel, stderrWriter, stderrChannel := machineEventLogger.CreateContainerOutputWriter()

	err := ExecuteCommand(client, compInfo, cmdArr, show, stdoutWriter, stderrWriter)

	// Close the writers and wait for an acknowledgement that the reader loop has exited (to ensure we get ALL container output)
	closeWriterAndWaitForAck(stdoutWriter, stdoutChannel, stderrWriter, stderrChannel)

	// Emit close event
	machineEventLogger.DevFileCommandExecutionComplete(exec.Id, exec.Component, command, convertGroupKindToString(exec), machineoutput.TimestampNow(), err)
	if err != nil {
		return errors.Wrapf(err, "unable to execute the build command")
	}

	s.End(true)

	return nil
}

// ExecuteDevfileRunAction executes the devfile run command action using the supervisord devrun program
func ExecuteDevfileRunAction(client ExecClient, exec common.Exec, commandName string, compInfo adaptersCommon.ComponentInfo, show bool, machineEventLogger machineoutput.MachineEventLoggingClient) error {
	var s *log.Status

	// Exec the supervisord ctl stop and start for the devrun program
	type devRunExecutable struct {
		command []string
	}
	devRunExecs := []devRunExecutable{
		{
			command: []string{adaptersCommon.SupervisordBinaryPath, adaptersCommon.SupervisordCtlSubCommand, "stop", "all"},
		},
		{
			command: []string{adaptersCommon.SupervisordBinaryPath, adaptersCommon.SupervisordCtlSubCommand, "start", string(adaptersCommon.DefaultDevfileRunCommand)},
		},
	}

	s = log.Spinnerf("Executing %s command %q", commandName, exec.CommandLine)
	defer s.End(false)

	for _, devRunExec := range devRunExecs {

		// Emit DevFileCommandExecutionBegin JSON event (if machine output logging is enabled)
		machineEventLogger.DevFileCommandExecutionBegin(exec.Id, exec.Component, exec.CommandLine, convertGroupKindToString(exec), machineoutput.TimestampNow())

		// Capture container text and log to the screen as JSON events (machine output only)
		stdoutWriter, stdoutChannel, stderrWriter, stderrChannel := machineEventLogger.CreateContainerOutputWriter()

		err := ExecuteCommand(client, compInfo, devRunExec.command, show, stdoutWriter, stderrWriter)

		// Close the writers and wait for an acknowledgement that the reader loop has exited (to ensure we get ALL container output)
		closeWriterAndWaitForAck(stdoutWriter, stdoutChannel, stderrWriter, stderrChannel)

		// Emit close event
		machineEventLogger.DevFileCommandExecutionComplete(exec.Id, exec.Component, exec.CommandLine, convertGroupKindToString(exec), machineoutput.TimestampNow(), err)
		if err != nil {
			return errors.Wrapf(err, "unable to execute the run command")
		}

	}
	s.End(true)

	return nil
}

// ExecuteDevfileRunActionWithoutRestart executes devfile run command without restarting.
func ExecuteDevfileRunActionWithoutRestart(client ExecClient, exec common.Exec, commandName string, compInfo adaptersCommon.ComponentInfo, show bool, machineEventLogger machineoutput.MachineEventLoggingClient) error {
	var s *log.Status

	type devRunExecutable struct {
		command []string
	}
	// with restart false, executing only supervisord start command, if the command is already running, supvervisord will not restart it.
	// if the command is failed or not running suprvisord would start it.
	devRunExec := devRunExecutable{
		command: []string{adaptersCommon.SupervisordBinaryPath, adaptersCommon.SupervisordCtlSubCommand, "start", string(adaptersCommon.DefaultDevfileRunCommand)},
	}

	s = log.Spinnerf("Executing %s command %q, if not running", commandName, exec.CommandLine)
	defer s.End(false)

	// Emit DevFileCommandExecutionBegin JSON event (if machine output logging is enabled)
	machineEventLogger.DevFileCommandExecutionBegin(exec.Id, exec.Component, exec.CommandLine, convertGroupKindToString(exec), machineoutput.TimestampNow())

	// Capture container text and log to the screen as JSON events (machine output only)
	stdoutWriter, stdoutChannel, stderrWriter, stderrChannel := machineEventLogger.CreateContainerOutputWriter()

	err := ExecuteCommand(client, compInfo, devRunExec.command, show, stdoutWriter, stderrWriter)

	// Close the writers and wait for an acknowledgement that the reader loop has exited (to ensure we get ALL container output)
	closeWriterAndWaitForAck(stdoutWriter, stdoutChannel, stderrWriter, stderrChannel)

	// Emit close event
	machineEventLogger.DevFileCommandExecutionComplete(exec.Id, exec.Component, exec.CommandLine, convertGroupKindToString(exec), machineoutput.TimestampNow(), err)
	if err != nil {
		return errors.Wrapf(err, "unable to execute the run command")
	}

	s.End(true)

	return nil
}

// ExecuteDevfileDebugAction executes the devfile debug command action using the supervisord debugrun program
func ExecuteDevfileDebugAction(client ExecClient, exec common.Exec, commandName string, compInfo adaptersCommon.ComponentInfo, show bool, machineEventLogger machineoutput.MachineEventLoggingClient) error {
	var s *log.Status

	// Exec the supervisord ctl stop and start for the debugRun program
	type debugRunExecutable struct {
		command []string
	}
	debugRunExecs := []debugRunExecutable{
		{
			command: []string{adaptersCommon.SupervisordBinaryPath, adaptersCommon.SupervisordCtlSubCommand, "stop", "all"},
		},
		{
			command: []string{adaptersCommon.SupervisordBinaryPath, adaptersCommon.SupervisordCtlSubCommand, "start", string(adaptersCommon.DefaultDevfileDebugCommand)},
		},
	}

	s = log.Spinnerf("Executing %s command %q", commandName, exec.CommandLine)
	defer s.End(false)

	for _, debugRunExec := range debugRunExecs {

		// Emit DevFileCommandExecutionBegin JSON event (if machine output logging is enabled)
		machineEventLogger.DevFileCommandExecutionBegin(exec.Id, exec.Component, exec.CommandLine, convertGroupKindToString(exec), machineoutput.TimestampNow())

		// Capture container text and log to the screen as JSON events (machine output only)
		stdoutWriter, stdoutChannel, stderrWriter, stderrChannel := machineEventLogger.CreateContainerOutputWriter()

		err := ExecuteCommand(client, compInfo, debugRunExec.command, show, stdoutWriter, stderrWriter)

		// Close the writers and wait for an acknowledgement that the reader loop has exited (to ensure we get ALL container output)
		closeWriterAndWaitForAck(stdoutWriter, stdoutChannel, stderrWriter, stderrChannel)

		// Emit close event
		machineEventLogger.DevFileCommandExecutionComplete(exec.Id, exec.Component, exec.CommandLine, convertGroupKindToString(exec), machineoutput.TimestampNow(), err)

		if err != nil {
			return errors.Wrapf(err, "unable to execute the run command")
		}
	}

	s.End(true)

	return nil
}

// ExecuteDevfileDebugActionWithoutRestart executes devfile run command without restarting.
func ExecuteDevfileDebugActionWithoutRestart(client ExecClient, exec common.Exec, commandName string, compInfo adaptersCommon.ComponentInfo, show bool, machineEventLogger machineoutput.MachineEventLoggingClient) error {
	var s *log.Status

	type devDebugExecutable struct {
		command []string
	}
	// with restart false, executing only supervisord start command, if the command is already running, supvervisord will not restart it.
	// if the command is failed or not running suprvisord would start it.
	devDebugExec := devDebugExecutable{
		command: []string{adaptersCommon.SupervisordBinaryPath, adaptersCommon.SupervisordCtlSubCommand, "start", string(adaptersCommon.DefaultDevfileDebugCommand)},
	}

	// Emit DevFileCommandExecutionBegin JSON event (if machine output logging is enabled)
	machineEventLogger.DevFileCommandExecutionBegin(exec.Id, exec.Component, exec.CommandLine, convertGroupKindToString(exec), machineoutput.TimestampNow())

	// Capture container text and log to the screen as JSON events (machine output only)
	stdoutWriter, stdoutChannel, stderrWriter, stderrChannel := machineEventLogger.CreateContainerOutputWriter()

	s = log.Spinnerf("Executing %s command %q, if not running", commandName, exec.CommandLine)
	defer s.End(false)

	err := ExecuteCommand(client, compInfo, devDebugExec.command, show, stdoutWriter, stderrWriter)

	// Close the writers and wait for an acknowledgement that the reader loop has exited (to ensure we get ALL container output)
	closeWriterAndWaitForAck(stdoutWriter, stdoutChannel, stderrWriter, stderrChannel)

	// Emit close event
	machineEventLogger.DevFileCommandExecutionComplete(exec.Id, exec.Component, exec.CommandLine, convertGroupKindToString(exec), machineoutput.TimestampNow(), err)

	if err != nil {
		return errors.Wrapf(err, "unable to execute the run command")
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
