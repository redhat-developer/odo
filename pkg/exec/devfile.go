package exec

import (
	"fmt"

	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/log"
)

// ExecuteDevfileBuildAction executes the devfile build command action
func ExecuteDevfileBuildAction(client ExecClient, action common.DevfileCommandAction, commandName, podName, containerName string, show bool) error {
	var s *log.Status

	// Change to the workdir and execute the command
	var cmdArr []string
	if action.Workdir != nil {
		cmdArr = []string{"/bin/sh", "-c", "cd " + *action.Workdir + " && " + *action.Command}
	} else {
		cmdArr = []string{"/bin/sh", "-c", *action.Command}
	}

	if show {
		s = log.SpinnerNoSpin("Executing " + commandName + " command " + fmt.Sprintf("%q", *action.Command))
	} else {
		s = log.Spinner("Executing " + commandName + " command " + fmt.Sprintf("%q", *action.Command))
	}

	defer s.End(false)

	err := ExecuteCommand(client, podName, containerName, cmdArr, show)
	if err != nil {
		return err
	}
	s.End(true)

	return nil
}

// ExecuteDevfileRunAction executes the devfile run command action using the supervisord devrun program
func ExecuteDevfileRunAction(client ExecClient, action common.DevfileCommandAction, commandName, podName, containerName string, show bool) error {
	var s *log.Status

	// Exec the supervisord ctl stop and start for the devrun program
	type devRunExecutable struct {
		command []string
	}
	devRunExecs := []devRunExecutable{
		{
			command: []string{adaptersCommon.SupervisordBinaryPath, "ctl", "stop", "all"},
		},
		{
			command: []string{adaptersCommon.SupervisordBinaryPath, "ctl", "start", string(adaptersCommon.DefaultDevfileRunCommand)},
		},
	}

	s = log.Spinner("Executing " + commandName + " command " + fmt.Sprintf("%q", *action.Command))
	defer s.End(false)

	for _, devRunExec := range devRunExecs {

		err := ExecuteCommand(client, podName, containerName, devRunExec.command, show)
		if err != nil {
			s.End(false)
			return err
		}
	}
	s.End(true)

	return nil
}
