package remotecmd

import (
	"errors"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// devfileCommandToRemoteCmdDefinition builds and returns a new remotecmd.CommandDefinition object from the specified devfileCmd.
// An error is returned for non-exec Devfile commands.
func DevfileCommandToRemoteCmdDefinition(devfileCmd devfilev1.Command) (CommandDefinition, error) {
	if devfileCmd.Exec == nil {
		return CommandDefinition{}, errors.New(" only Exec commands are supported")
	}

	envVars := make([]CommandEnvVar, 0, len(devfileCmd.Exec.Env))
	for _, e := range devfileCmd.Exec.Env {
		envVars = append(envVars, CommandEnvVar{Key: e.Name, Value: e.Value})
	}

	return CommandDefinition{
		Id:         devfileCmd.Id,
		WorkingDir: devfileCmd.Exec.WorkingDir,
		EnvVars:    envVars,
		CmdLine:    devfileCmd.Exec.CommandLine,
	}, nil
}
