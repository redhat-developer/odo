package common

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/glog"

	"github.com/openshift/odo/pkg/devfile/versions"
	"github.com/openshift/odo/pkg/devfile/versions/common"
)

const (
	DefaultDevfileBuildCommand = "devBuild"
	DefaultDevfileRunCommand   = "devRun"
	DefaultDevfileDebugCommand = "devDebug"
	DefaultDevfileTestCommand  = "devTest"
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
func GetCommand(data versions.DevfileData, commandName string) (command common.DevfileCommand) {
	var supportedCommand common.DevfileCommand

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

// GetSupportedCommands iterates through the devfile commands in the devfile and returns
// commands with 1. odo supported command name 2. odo supported command actions
func GetSupportedCommands(data versions.DevfileData) []common.DevfileCommand {
	var supportedCommands []common.DevfileCommand

	for _, command := range data.GetCommands() {
		var supportedCommand common.DevfileCommand

		// Check if the command is supported by default
		if IsDevfileCommandSupported(command.Name) {

			// Get the supported command actions
			supportedCommandActions := getSupportedCommandActions(command)

			// if there is a supported command action of type exec, save it
			if len(supportedCommandActions) > 0 {
				supportedCommand.Name = command.Name
				supportedCommand.Actions = supportedCommandActions
				supportedCommands = append(supportedCommands, supportedCommand)
			}
		}
	}

	return supportedCommands
}

// getSupportedCommandActions returns the supported command action
// 1. action has to be of type exec 2. component should be present
func getSupportedCommandActions(command common.DevfileCommand) (supportedCommandActions []common.DevfileCommandAction) {
	for _, action := range command.Actions {
		// Check if the command action is of type exec
		if *action.Type == common.DevfileCommandTypeExec && *action.Component != "" {
			glog.V(3).Infof("Found command %v for component %v and type %v", command.Name, *action.Component, *action.Type)
			supportedCommandActions = append(supportedCommandActions, action)
		}
	}

	return
}

// GetRunCommandComponents iterates through the components in the devfile and returns a slice of the corresponding containers
func GetRunCommandComponents(data versions.DevfileData, devfileRunCmd string) []string {
	var components []string
	var emptyCommand common.DevfileCommand

	if devfileRunCmd != "" {
		command := GetCommand(data, devfileRunCmd)
		if !reflect.DeepEqual(emptyCommand, command) {
			for _, action := range command.Actions {
				components = append(components, *action.Component)
			}
		}
	} else {
		for _, command := range GetSupportedCommands(data) {
			if strings.Contains(command.Name, DefaultDevfileRunCommand) {
				for _, action := range command.Actions {
					components = append(components, *action.Component)
				}
			}
		}
	}

	return components
}

// ValidateAndGetPushDevfileCommands returns a build and run command. It checks if a build
// or a run command has been explicitly provided during odo push, otherwise it
// iterates through the devfile commands and returns the default supported command. If neither
// a build command nor a run command is found, it throws an error
func ValidateAndGetPushDevfileCommands(data versions.DevfileData, devfileBuildCmd, devfileRunCmd string) ([]common.DevfileCommand, error) {
	var pushDevfileCommands []common.DevfileCommand
	var emptyCommand common.DevfileCommand
	validateBuildCommand := false
	validateRunCommand := false

	devfileSupportedCommands := GetSupportedCommands(data)

	// Validate the build command if it was provided during odo push
	if devfileBuildCmd != "" {
		devfileCommand := GetCommand(data, devfileBuildCmd)
		if !reflect.DeepEqual(emptyCommand, devfileCommand) {
			pushDevfileCommands = append(pushDevfileCommands, devfileCommand)
			validateBuildCommand = true
		}
	}

	// Validate the run command if it was provided during odo push
	if devfileRunCmd != "" {
		devfileCommand := GetCommand(data, devfileRunCmd)
		if !reflect.DeepEqual(emptyCommand, devfileCommand) {
			pushDevfileCommands = append(pushDevfileCommands, devfileCommand)
			validateRunCommand = true
		}
	}

	// If neither is validated, iterate the devfile commands to see for a list of supported commands
	for _, supportedCommand := range devfileSupportedCommands {
		if reflect.DeepEqual(supportedCommand.Name, DefaultDevfileBuildCommand) && !validateBuildCommand {
			pushDevfileCommands = append(pushDevfileCommands, supportedCommand)
			validateBuildCommand = true
		}
		if reflect.DeepEqual(supportedCommand.Name, DefaultDevfileRunCommand) && !validateRunCommand {
			pushDevfileCommands = append(pushDevfileCommands, supportedCommand)
			validateRunCommand = true
		}
	}

	if !validateBuildCommand || !validateRunCommand {
		return []common.DevfileCommand{}, fmt.Errorf("devfile command validation failed, validateBuildCommand: %v validateRunCommand: %v", validateBuildCommand, validateRunCommand)
	}

	return pushDevfileCommands, nil
}

// IsDevfileCommandSupported checks if a devfile command is supported by default
func IsDevfileCommandSupported(commandName string) bool {
	isSupported := false

	switch commandName {
	case DefaultDevfileBuildCommand:
		fallthrough
	case DefaultDevfileRunCommand:
		fallthrough
	case DefaultDevfileDebugCommand:
		fallthrough
	case DefaultDevfileTestCommand:
		isSupported = true
	default:
		isSupported = false
	}

	return isSupported
}
