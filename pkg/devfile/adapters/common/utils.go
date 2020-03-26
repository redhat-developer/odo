package common

import (
	"fmt"
	"reflect"

	"github.com/golang/glog"

	"github.com/openshift/odo/pkg/devfile/versions"
	"github.com/openshift/odo/pkg/devfile/versions/common"
)

type PredefinedDevfileCommands string

const (
	DefaultDevfileBuildCommand PredefinedDevfileCommands = "devbuild"
	DefaultDevfileRunCommand   PredefinedDevfileCommands = "devrun"
)

// GetSupportedComponents iterates through the components in the devfile and returns a list of odo supported components
func GetSupportedComponents(data versions.DevfileData) []common.DevfileComponent {
	var components []common.DevfileComponent
	// Only components with aliases are considered because without an alias commands cannot reference them
	for _, comp := range data.GetAliasedComponents() {
		// Currently odo only uses devfile components of type dockerimage, since most of the Che registry devfiles use it
		if comp.Type == common.DevfileComponentTypeDockerimage {
			glog.V(3).Infof("Found component %v with alias %v\n", comp.Type, *comp.Alias)
			components = append(components, comp)
		}
	}
	return components
}

// GetCommand iterates through the devfile commands and returns the associated devfile command
func GetCommand(data versions.DevfileData, commandName string) (supportedCommand common.DevfileCommand) {
	for _, command := range data.GetCommands() {
		if command.Name == commandName {

			// Get the supported command actions
			supportedCommandActions := getSupportedCommandActions(command)

			// if there is a supported command action of type exec, save it
			if len(supportedCommandActions) > 0 {
				supportedCommand.Name = command.Name
				supportedCommand.Actions = supportedCommandActions
			}

			return supportedCommand
		}
	}

	return
}

// getSupportedCommandActions returns the supported command action
func getSupportedCommandActions(command common.DevfileCommand) (supportedCommandActions []common.DevfileCommandAction) {
	glog.V(3).Infof("Validating command's action for %v ", command.Name)
	for _, action := range command.Actions {
		// Check if the command action is of type exec
		if validateAction(action) {
			glog.V(3).Infof("Found command %v for component %v", command.Name, *action.Component)
			supportedCommandActions = append(supportedCommandActions, action)
		}
	}

	return
}

// validateAction validates the given action
// 1. action has to be of type exec 2. component should be present
// 3. command should be present
func validateAction(action common.DevfileCommandAction) bool {
	if *action.Type != common.DevfileCommandTypeExec {
		return false
	}

	if action.Component == nil || *action.Component == "" {
		return false
	}

	if action.Command == nil || *action.Command == "" {
		return false
	}

	return true
}

// GetBuildCommand iterates through the components in the devfile and returns the build command
func GetBuildCommand(data versions.DevfileData, devfileBuildCmd string) (buildCommand common.DevfileCommand) {
	if devfileBuildCmd != "" {
		buildCommand = GetCommand(data, devfileBuildCmd)
	} else {
		buildCommand = GetCommand(data, string(DefaultDevfileBuildCommand))
	}

	return
}

// GetRunCommand iterates through the components in the devfile and returns the run command
func GetRunCommand(data versions.DevfileData, devfileRunCmd string) (runCommand common.DevfileCommand) {
	if devfileRunCmd != "" {
		runCommand = GetCommand(data, devfileRunCmd)
	} else {
		runCommand = GetCommand(data, string(DefaultDevfileRunCommand))
	}

	return
}

// IsCommandPresent checks if the given command is empty or not
func IsCommandPresent(command common.DevfileCommand) bool {
	var emptyCommand common.DevfileCommand
	isPresent := false

	if !reflect.DeepEqual(emptyCommand, command) {
		isPresent = true
	}

	return isPresent
}

// ValidateAndGetPushDevfileCommands validates the build and the run command,
// if provided through odo push or else checks the devfile for devBuild and devRun.
// It returns the build and run commands if its validated successfully, error otherwise.
func ValidateAndGetPushDevfileCommands(data versions.DevfileData, devfileBuildCmd, devfileRunCmd string) ([]common.DevfileCommand, error) {
	var pushDevfileCommands []common.DevfileCommand
	// var emptyCommand common.DevfileCommand
	validateBuildCommand, validateRunCommand := false, false

	buildCommand := GetBuildCommand(data, devfileBuildCmd)
	if devfileBuildCmd == "" && !IsCommandPresent(buildCommand) {
		// If there is no build command either in the devfile or through odo push, default validate to true since build command is optional
		validateBuildCommand = true
		glog.V(3).Infof("No Build command was provided")
	} else if IsCommandPresent(buildCommand) && IsCommandValid(data, buildCommand) {
		// If the build command is present, validate it
		pushDevfileCommands = append(pushDevfileCommands, buildCommand)
		validateBuildCommand = true
		glog.V(3).Infof("Build command %v validated", buildCommand.Name)
	}

	runCommand := GetRunCommand(data, devfileRunCmd)
	if IsCommandPresent(runCommand) && IsCommandValid(data, runCommand) {
		// If the run command is present, validate it
		pushDevfileCommands = append(pushDevfileCommands, runCommand)
		validateRunCommand = true
		glog.V(3).Infof("Run command %v validated", runCommand.Name)
	}

	if !validateBuildCommand || !validateRunCommand {
		return []common.DevfileCommand{}, fmt.Errorf("devfile command validation failed, validateBuildCommand: %v validateRunCommand: %v", validateBuildCommand, validateRunCommand)
	}

	return pushDevfileCommands, nil
}

// IsCommandValid checks if a command is valid. It checks
// 1. if the command references a component that is present
// 2. if the referenced component is of type dockerimage
func IsCommandValid(data versions.DevfileData, command common.DevfileCommand) bool {

	var isCommandValid bool
	var isCommandActionValid []bool
	isCommandActionValid = make([]bool, len(command.Actions))

	// GetSupportedComponents gets components of type dockerimage
	components := GetSupportedComponents(data)

	for i, action := range command.Actions {
		isCommandActionValid[i] = false
		for _, component := range components {
			if *action.Component == *component.Alias {
				isCommandActionValid[i] = true
			}
		}
	}

	// if any of the command action is invalid, the command is invalid
	for _, isActionValid := range isCommandActionValid {
		isCommandValid = isActionValid
		if !isActionValid {
			break
		}
	}

	return isCommandValid
}
