package common

import (
	"fmt"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/pkg/errors"
)

type simpleCommand struct {
	info        ComponentInfo
	adapter     commandExecutor
	cmd         []string
	id          string
	component   string
	originalCmd string
	group       string
}

func newSimpleCommand(command common.DevfileCommand, executor commandExecutor) (command, error) {
	exe := command.Exec

	// deal with environment variables
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
	var cmd []string
	if exe.WorkingDir != "" {
		// since we are using /bin/sh -c, the command needs to be within a single double quote instance, for example "cd /tmp && pwd"
		cmd = []string{ShellExecutable, "-c", "cd " + exe.WorkingDir + " && " + cmdLine}
	} else {
		cmd = []string{ShellExecutable, "-c", cmdLine}
	}

	// create the component info associated with the command
	info, err := executor.ComponentInfo(command)
	if err != nil {
		return nil, err
	}

	return simpleCommand{
		info:        info,
		adapter:     executor,
		cmd:         cmd,
		id:          command.GetID(),
		component:   exe.Component,
		originalCmd: exe.CommandLine,
		group:       convertGroupKindToString(exe),
	}, nil
}

func (s simpleCommand) Execute(show bool) error {
	msg := fmt.Sprintf("Executing %s command %q, if not running", s.id, s.originalCmd)
	spinner := log.ExplicitSpinner(msg, show)
	defer spinner.End(false)

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

	spinner.End(true)

	return nil
}
