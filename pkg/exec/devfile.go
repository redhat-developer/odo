package exec

import (
	"fmt"

	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/log"
	"github.com/pkg/errors"
)

// ExecuteDevfileBuildAction executes the devfile build command action
func ExecuteDevfileBuildAction(client ExecClient, exec common.Exec, commandName string, compInfo adaptersCommon.ComponentInfo, show bool) error {
	var s *log.Status

	// Change to the workdir and execute the command
	var cmdArr []string
	if exec.WorkingDir != nil {
		// since we are using /bin/sh -c, the command needs to be within a single double quote instance, for example "cd /tmp && pwd"
		cmdArr = []string{adaptersCommon.ShellExecutable, "-c", "cd " + *exec.WorkingDir + " && " + exec.CommandLine}
	} else {
		cmdArr = []string{adaptersCommon.ShellExecutable, "-c", exec.CommandLine}
	}

	if show {
		s = log.SpinnerNoSpin("Executing " + commandName + " command " + fmt.Sprintf("%q", exec.CommandLine))
	} else {
		s = log.Spinnerf("Executing %s command %q", commandName, exec.CommandLine)
	}

	defer s.End(false)

	err := ExecuteCommand(client, compInfo, cmdArr, show)
	if err != nil {
		return errors.Wrapf(err, "unable to execute the build command")
	}
	s.End(true)

	return nil
}

// ExecuteDevfileRunAction executes the devfile run command action using the supervisord devrun program
func ExecuteDevfileRunAction(client ExecClient, exec common.Exec, commandName string, compInfo adaptersCommon.ComponentInfo, show bool) error {
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

		err := ExecuteCommand(client, compInfo, devRunExec.command, show)
		if err != nil {
			return errors.Wrapf(err, "unable to execute the run command")
		}
	}
	s.End(true)

	return nil
}

// ExecuteDevfileRunActionWithoutRestart executes devfile run command without restarting.
func ExecuteDevfileRunActionWithoutRestart(client ExecClient, exec common.Exec, commandName string, compInfo adaptersCommon.ComponentInfo, show bool) error {
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

	err := ExecuteCommand(client, compInfo, devRunExec.command, show)
	if err != nil {
		return errors.Wrapf(err, "unable to execute the run command")
	}

	s.End(true)

	return nil
}
