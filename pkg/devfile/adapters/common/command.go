package common

import (
	"fmt"
	"reflect"

	"github.com/openshift/odo/pkg/devfile/parser/data"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"k8s.io/klog"
)

// GetCommand iterates through the devfile commands and returns the associated devfile command
func getCommand(data data.DevfileData, commandName string, required bool) (supportedCommand common.DevfileCommand, err error) {
	for _, command := range data.GetCommands() {
		if command.Exec.Id == commandName {

			// Get the supported actions
			err = validateCommand(data, command)

			if err != nil {
				return common.DevfileCommand{}, err
			}
			// The command is supported, use it
			supportedCommand.Exec = command.Exec
			return supportedCommand, nil
		}
	}

	// The command was not found
	msg := fmt.Sprintf("The command \"%v\" was not found in the devfile", commandName)
	if required {
		// Not found and required, return an error
		err = fmt.Errorf(msg)
	} else {
		// Not found and optional, so just log it
		klog.V(3).Info(msg)
	}

	return
}

// validateCommand validates the given command
// 1. command has to be of type exec
// 2. component should be present
// 3. command should be present
func validateCommand(data data.DevfileData, command common.DevfileCommand) (err error) {

	// type must be exec
	if command.Type != common.ExecCommandType {
		return fmt.Errorf("Actions must be of type \"exec\"")
	}

	// component must be specified
	if &command.Exec.Component == nil || command.Exec.Component == "" {
		return fmt.Errorf("Actions must reference a component")
	}

	// must specify a command
	if &command.Exec.CommandLine == nil || command.Exec.CommandLine == "" {
		return fmt.Errorf("Actions must have a command")
	}

	// must map to a supported component
	components := GetSupportedComponents(data)

	isActionValid := false
	for _, component := range components {
		if command.Exec.Component == component.Container.Name && isComponentSupported(component) {
			isActionValid = true
		}
	}
	if !isActionValid {
		return fmt.Errorf("the command does not map to a supported component")
	}

	return
}

// GetInitCommand iterates through the components in the devfile and returns the init command
func GetInitCommand(data data.DevfileData, devfileInitCmd string) (initCommand common.DevfileCommand, err error) {
	if devfileInitCmd != "" {
		// a init command was specified so if it is not found then it is an error
		return getCommand(data, devfileInitCmd, true)
	}
	// a init command was not specified so if it is not found then it is not an error
	return getCommand(data, string(DefaultDevfileInitCommand), false)
}

// GetBuildCommand iterates through the components in the devfile and returns the build command
func GetBuildCommand(data data.DevfileData, devfileBuildCmd string) (buildCommand common.DevfileCommand, err error) {
	if devfileBuildCmd != "" {
		// a build command was specified so if it is not found then it is an error
		return getCommand(data, devfileBuildCmd, true)
	}
	// a build command was not specified so if it is not found then it is not an error
	return getCommand(data, string(DefaultDevfileBuildCommand), false)
}

// GetRunCommand iterates through the components in the devfile and returns the run command
func GetRunCommand(data data.DevfileData, devfileRunCmd string) (runCommand common.DevfileCommand, err error) {
	if devfileRunCmd != "" {
		return getCommand(data, devfileRunCmd, true)
	}
	return getCommand(data, string(DefaultDevfileRunCommand), true)
}

// ValidateAndGetPushDevfileCommands validates the build and the run command,
// if provided through odo push or else checks the devfile for devBuild and devRun.
// It returns the build and run commands if its validated successfully, error otherwise.
func ValidateAndGetPushDevfileCommands(data data.DevfileData, devfileInitCmd, devfileBuildCmd, devfileRunCmd string) (pushDevfileCommands []common.DevfileCommand, err error) {
	var emptyCommand common.DevfileCommand
	isInitCommandValid, isBuildCommandValid, isRunCommandValid := false, false, false

	initCommand, initCmdErr := GetInitCommand(data, devfileInitCmd)

	isInitCmdEmpty := reflect.DeepEqual(emptyCommand, initCommand)
	if isInitCmdEmpty && initCmdErr == nil {
		// If there was no init command specified through odo push and no default init command in the devfile, default validate to true since the init command is optional
		isInitCommandValid = true
		klog.V(3).Infof("No init command was provided")
	} else if !isInitCmdEmpty && initCmdErr == nil {
		isInitCommandValid = true
		pushDevfileCommands = append(pushDevfileCommands, initCommand)
		klog.V(3).Infof("Init command: %v", initCommand.Exec.Id)
	}

	buildCommand, buildCmdErr := GetBuildCommand(data, devfileBuildCmd)

	isBuildCmdEmpty := reflect.DeepEqual(emptyCommand, buildCommand)
	if isBuildCmdEmpty && buildCmdErr == nil {
		// If there was no build command specified through odo push and no default build command in the devfile, default validate to true since the build command is optional
		isBuildCommandValid = true
		klog.V(3).Infof("No build command was provided")
	} else if !reflect.DeepEqual(emptyCommand, buildCommand) && buildCmdErr == nil {
		isBuildCommandValid = true
		pushDevfileCommands = append(pushDevfileCommands, buildCommand)
		klog.V(3).Infof("Build command: %v", buildCommand.Exec.Id)
	}

	runCommand, runCmdErr := GetRunCommand(data, devfileRunCmd)
	if runCmdErr == nil && !reflect.DeepEqual(emptyCommand, runCommand) {
		pushDevfileCommands = append(pushDevfileCommands, runCommand)
		isRunCommandValid = true
		klog.V(3).Infof("Run command: %v", runCommand.Exec.Id)
	}

	// If either command had a problem, return an empty list of commands and an error
	if !isInitCommandValid || !isBuildCommandValid || !isRunCommandValid {
		commandErrors := ""
		if initCmdErr != nil {
			commandErrors += fmt.Sprintf(initCmdErr.Error(), "\n")
		}
		if buildCmdErr != nil {
			commandErrors += fmt.Sprintf(buildCmdErr.Error(), "\n")
		}
		if runCmdErr != nil {
			commandErrors += fmt.Sprintf(runCmdErr.Error(), "\n")
		}
		return []common.DevfileCommand{}, fmt.Errorf(commandErrors)
	}

	return pushDevfileCommands, nil
}
