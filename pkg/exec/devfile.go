package exec

import (
	"fmt"
	"io"
	"k8s.io/klog"

	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/pkg/errors"
)

// ExecuteDevfileCommandSynchronously executes the devfile init, build and test command actions synchronously
func ExecuteDevfileCommandSynchronously(client ExecClient, exec common.Exec, compInfo adaptersCommon.ComponentInfo, show bool, machineEventLogger machineoutput.MachineEventLoggingClient) error {
	// Change to the workdir and execute the command
	var cmd executable
	if exec.WorkingDir != "" {
		// since we are using /bin/sh -c, the command needs to be within a single double quote instance, for example "cd /tmp && pwd"
		cmd = executable{adaptersCommon.ShellExecutable, "-c", "cd " + exec.WorkingDir + " && " + exec.CommandLine}
	} else {
		cmd = executable{adaptersCommon.ShellExecutable, "-c", exec.CommandLine}
	}
	return Execute(client, exec, compInfo, show, machineEventLogger, []executable{cmd})
}

// DefaultCommands returns the devfile commands to execute based on the specified command options
func DefaultCommands(debug, restart bool) []executable {
	cmd := string(adaptersCommon.DefaultDevfileRunCommand)
	if debug {
		cmd = string(adaptersCommon.DefaultDevfileDebugCommand)
	}
	klog.V(4).Infof("restart:false, not restarting %s", cmd)

	// with restart false, executing only supervisord start command, if the command is already running, supvervisord will not restart it.
	// if the command is failed or not running supervisord would start it.
	execs := []executable{
		{adaptersCommon.SupervisordBinaryPath, adaptersCommon.SupervisordCtlSubCommand, "start", cmd},
	}

	if restart {
		// first stop any running command
		stopSupervisorExec := executable{adaptersCommon.SupervisordBinaryPath, adaptersCommon.SupervisordCtlSubCommand, "stop", "all"}
		execs = append([]executable{stopSupervisorExec}, execs...)
	}
	return execs
}

type executable []string

// Execute executes the specified devfile exec command appropriately wrapped into the appropriate sequence of executable, usually by calling DefaultCommands
func Execute(client ExecClient, exec common.Exec, compInfo adaptersCommon.ComponentInfo, show bool, machineEventLogger machineoutput.MachineEventLoggingClient, subcommands []executable) error {
	msg := fmt.Sprintf("Executing %s command %q, if not running", exec.Id, exec.CommandLine)
	var s *log.Status
	if show {
		s = log.SpinnerNoSpin(msg)
	} else {
		s = log.Spinnerf(msg)
	}
	defer s.End(false)

	for _, subcommand := range subcommands {

		// Emit DevFileCommandExecutionBegin JSON event (if machine output logging is enabled)
		machineEventLogger.DevFileCommandExecutionBegin(exec.Id, exec.Component, exec.CommandLine, convertGroupKindToString(exec), machineoutput.TimestampNow())

		// Capture container text and log to the screen as JSON events (machine output only)
		stdoutWriter, stdoutChannel, stderrWriter, stderrChannel := machineEventLogger.CreateContainerOutputWriter()

		err := ExecuteCommand(client, compInfo, subcommand, show, stdoutWriter, stderrWriter)

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
