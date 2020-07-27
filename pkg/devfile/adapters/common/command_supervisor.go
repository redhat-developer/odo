package common

import "github.com/openshift/odo/pkg/machineoutput"

type supervisorCommand struct {
	adapter commandExecutor
	cmd     []string
	info    ComponentInfo
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
