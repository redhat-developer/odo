package common

import (
	"fmt"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
)

// execCommand is a command implementation for non-composite commands
type execCommand struct {
	info        ComponentInfo
	adapter     commandExecutor
	cmd         []string // cmd represents the effective command that will be run in the container
	id          string
	component   string
	originalCmd string // originalCmd records the command as defined in the devfile
	group       string
	msg         string
}

// newExecCommand creates a new execCommand instance, adapting the devfile-defined command to run in the target component's
// container, modifying it to add environment variables or adapting the path as needed.
func newExecCommand(command devfilev1.Command, executor commandExecutor) (command, error) {
	exe := command.Exec

	// deal with environment variables
	var cmdLine string
	setEnvVariable := util.GetCommandStringFromEnvs(exe.Env)

	if setEnvVariable == "" {
		cmdLine = exe.CommandLine
	} else {
		cmdLine = setEnvVariable + " && " + exe.CommandLine
	}

	// Change to the workdir and execute the command
	var cmd []string
	if exe.WorkingDir != "" {
		// since we are using /bin/sh -c, the command needs to be within a single double quote instance, for example "cd /tmp && pwd"
		cmd = []string{ShellExecutable, "-c", "cd " + exe.WorkingDir + " && " + cmdLine}
	} else {
		cmd = []string{ShellExecutable, "-c", cmdLine}
	}

	return newOverriddenExecCommand(command, executor, cmd)
}

// newOverriddenExecCommand creates a new execCommand albeit overriding the command specified in the devfile with the specified one
// returning a pointer to the newly created instance so that clients can further modify it if needed.
// Note that the specified command will be run as-is in the target component's container so needs to be set accordingly as
// opposed to the implementation provided by newExecCommand which will take the devfile's command definition and adapt it to
// run in the container.
func newOverriddenExecCommand(command devfilev1.Command, executor commandExecutor, cmd []string) (*execCommand, error) {
	// create the component info associated with the command
	info, err := executor.ComponentInfo(command)
	if err != nil {
		return nil, err
	}

	originalCmd := command.Exec.CommandLine
	return &execCommand{
		info:        info,
		adapter:     executor,
		cmd:         cmd,
		id:          command.Id,
		component:   command.Exec.Component,
		originalCmd: originalCmd,
		group:       convertGroupKindToString(command.Exec),
		msg:         fmt.Sprintf("Executing %s command %q", command.Id, originalCmd),
	}, nil
}

func (s execCommand) Execute(show bool) error {
	var spinner *log.Status
	showSpinner := len(s.msg) > 0
	if showSpinner {
		spinner = log.ExplicitSpinner(s.msg, show)
		defer spinner.End(false)
	}

	// Emit DevFileCommandExecutionBegin JSON event (if machine output logging is enabled)
	logger := s.adapter.Logger()
	logger.DevFileCommandExecutionBegin(s.id, s.component, s.originalCmd, s.group, machineoutput.TimestampNow())

	// Capture container text and log to the screen as JSON events (machine output only)
	stdoutWriter, stdoutChannel, stderrWriter, stderrChannel := logger.CreateContainerOutputWriter()

	err := ExecuteCommand(s.adapter, s.info, s.cmd, show, stdoutWriter, stderrWriter)

	// Close the writers and wait for an acknowledgement that the reader loop has exited (to ensure we get ALL container output)
	closeWriterAndWaitForAck(stdoutWriter, stdoutChannel, stderrWriter, stderrChannel)

	// Emit close event
	logger.DevFileCommandExecutionComplete(s.id, s.component, s.originalCmd, s.group, machineoutput.TimestampNow(), err)

	if err != nil {
		return errors.Wrapf(err, "unable to execute the run command")
	}

	if showSpinner {
		spinner.End(true)
	}

	return nil
}
