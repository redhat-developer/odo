package common

import (
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/machineoutput"
)

type supervisorCommand struct {
	adapter commandExecutor
	cmd     []string
	info    ComponentInfo
}

func newSupervisorInitCommand(command common.DevfileCommand, adapter commandExecutor) (command, error) {
	cmd := []string{"-c", SupervisordConfFile, "-d"}
	info, err := adapter.SupervisorComponentInfo(command)
	return newErrorCheckedSupervisorCommand(err, adapter, info, cmd)
}

func newErrorCheckedSupervisorCommand(err error, adapter commandExecutor, info ComponentInfo, cmd []string) (command, error) {
	if err != nil {
		adapter.Logger().ReportError(err, machineoutput.TimestampNow())
		return nil, err
	}
	if !info.IsEmpty() {
		return newSupervisorCommand(cmd, info, adapter)
	}
	return nil, nil
}

func newSupervisorStopCommand(command common.DevfileCommand, adapter commandExecutor) (command, error) {
	cmd := []string{SupervisordCtlSubCommand, "stop", "all"}
	info, err := adapter.ComponentInfo(command)
	return newErrorCheckedSupervisorCommand(err, adapter, info, cmd)
}

func newSupervisorStartCommand(command common.DevfileCommand, cmd string, adapter commandExecutor) (command, error) {
	cmdLine := []string{SupervisordCtlSubCommand, "start", cmd}
	info, err := adapter.ComponentInfo(command)
	return newErrorCheckedSupervisorCommand(err, adapter, info, cmdLine)
}

func newSupervisorCommand(cmd []string, info ComponentInfo, adapter commandExecutor) (command, error) {
	// prepend supervisor binary path if command doesn't already start with it
	if cmd[0] != SupervisordBinaryPath {
		cmd = append([]string{SupervisordBinaryPath}, cmd...)
	}
	return supervisorCommand{
		adapter: adapter,
		cmd:     cmd,
		info:    info,
	}, nil
}

func (s supervisorCommand) Execute(show bool) error {
	err := ExecuteCommand(s.adapter, s.info, s.cmd, true, nil, nil)
	if err != nil {
		s.adapter.Logger().ReportError(err, machineoutput.TimestampNow())
		return err
	}
	return nil
}
