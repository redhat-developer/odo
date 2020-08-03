package common

import (
	"fmt"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/machineoutput"
)

type supervisorCommand struct {
	adapter commandExecutor
	cmd     []string
	info    ComponentInfo
}

func newSupervisorInitCommand(command common.DevfileCommand, adapter commandExecutor) (command, error) {
	cmd := []string{SupervisordBinaryPath, "-c", SupervisordConfFile, "-d"}
	info, err := adapter.SupervisorComponentInfo(command)
	if err != nil {
		adapter.Logger().ReportError(err, machineoutput.TimestampNow())
		return nil, err
	}
	if !info.IsEmpty() {
		return supervisorCommand{
			adapter: adapter,
			cmd:     cmd,
			info:    info,
		}, nil
	}
	return nil, nil
}

func newSupervisorStopCommand(command common.DevfileCommand, executor commandExecutor) (command, error) {
	cmd := []string{SupervisordBinaryPath, SupervisordCtlSubCommand, "stop", "all"}
	if stop, err := newOverridenSimpleCommand(command, executor, cmd); err == nil {
		// override spinner message
		stop.msg = fmt.Sprintf("Stopping %s command %q, if running", command.GetID(), command.Exec.CommandLine)
		return stop, err
	} else {
		return nil, err
	}
}

func newSupervisorStartCommand(command common.DevfileCommand, cmd string, adapter commandExecutor) (command, error) {
	cmdLine := []string{SupervisordBinaryPath, SupervisordCtlSubCommand, "start", cmd}
	return newOverridenSimpleCommand(command, adapter, cmdLine)
}

func (s supervisorCommand) Execute(show bool) error {
	err := ExecuteCommand(s.adapter, s.info, s.cmd, true, nil, nil)
	if err != nil {
		s.adapter.Logger().ReportError(err, machineoutput.TimestampNow())
		return err
	}
	return nil
}
