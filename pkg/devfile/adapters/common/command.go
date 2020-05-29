package common

import (
	"fmt"
	"reflect"

	"github.com/openshift/odo/pkg/devfile/parser/data"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"k8s.io/klog"
)

// GetCommand iterates through the devfile commands and returns the associated devfile command
func getCommand(data data.DevfileData, commandName string, groupType common.DevfileCommandGroupType) (supportedCommand common.DevfileCommand, err error) {

	commands := data.GetCommands()

	for _, command := range commands {

		command = updateGroupforCustomCommand(commandName, groupType, command)

		// validate command
		err = validateCommand(data, command)

		if err != nil {
			return common.DevfileCommand{}, err
		}

		// if command is specified via flags, it has the highest priority
		// search through all commands to find the specified command name
		// if not found fallback to error.
		if commandName != "" {

			if command.Exec.Id == commandName {

				if command.Exec.Group.Kind == "" {
					// Devfile V1 for commands passed from flags
					// Group type is not updated during conversion
					command.Exec.Group.Kind = groupType
				}

				// we have found the command with name, its groupType Should match to the flag
				// e.g --build-command "mybuild"
				// exec:
				//   id: mybuild
				// group:
				//   kind: build
				if command.Exec.Group.Kind != groupType {
					return supportedCommand, fmt.Errorf("mismatched group kind, command %s is of group kind %v groupType in devfile", commandName, groupType)

				}
				supportedCommand = command
				return supportedCommand, nil
			}
			continue
		}

		// if not command specified via flag, default command has the highest priority
		// We need to scan all the commands to find default command
		if command.Exec.Group.Kind == groupType && command.Exec.Group.IsDefault {
			supportedCommand = command
			return supportedCommand, nil
		}
	}

	if commandName == "" {
		// if default command is not found return the first command found for the matching type.
		for _, command := range commands {
			if command.Exec.Group.Kind == groupType {
				supportedCommand = command
				return supportedCommand, nil
			}

		}
	}

	// if any command specified via flag is not found in devfile then it is an error.
	if commandName != "" {
		err = fmt.Errorf("The command \"%v\" is not found in the devfile", commandName)
	} else {
		msg := fmt.Sprintf("The command type \"%v\" is not found in the devfile", groupType)
		// if run command is not found in devfile then it is an error
		if groupType == common.RunCommandGroupType {
			err = fmt.Errorf(msg)
		} else {
			klog.V(3).Info(msg)

		}
	}

	return
}

// validateCommand validates the given command
// 1. command has to be of type exec
// 2. component should be present
// 3. command should be present
// 4. command must have group
func validateCommand(data data.DevfileData, command common.DevfileCommand) (err error) {

	// type must be exec
	if command.Exec == nil {
		return fmt.Errorf("Command must be of type \"exec\"")
	}

	// component must be specified
	if command.Exec.Component == "" {
		return fmt.Errorf("Exec commands must reference a component")
	}

	// must specify a command
	if command.Exec.CommandLine == "" {
		return fmt.Errorf("Exec commands must have a command")
	}

	if command.Exec.Group == nil {
		return fmt.Errorf("Exec commands must have group")
	}

	// must map to a supported component
	components := GetSupportedComponents(data)

	isComponentValid := false
	for _, component := range components {
		if command.Exec.Component == component.Container.Name {
			isComponentValid = true
		}
	}
	if !isComponentValid {
		return fmt.Errorf("the command does not map to a supported component")
	}

	return
}

// GetInitCommand iterates through the components in the devfile and returns the init command
func GetInitCommand(data data.DevfileData, devfileInitCmd string) (initCommand common.DevfileCommand, err error) {

	return getCommand(data, devfileInitCmd, common.InitCommandGroupType)
}

// GetBuildCommand iterates through the components in the devfile and returns the build command
func GetBuildCommand(data data.DevfileData, devfileBuildCmd string) (buildCommand common.DevfileCommand, err error) {

	return getCommand(data, devfileBuildCmd, common.BuildCommandGroupType)
}

// GetRunCommand iterates through the components in the devfile and returns the run command
func GetRunCommand(data data.DevfileData, devfileRunCmd string) (runCommand common.DevfileCommand, err error) {

	return getCommand(data, devfileRunCmd, common.RunCommandGroupType)
}

// ValidateAndGetPushDevfileCommands validates the build and the run command,
// if provided through odo push or else checks the devfile for devBuild and devRun.
// It returns the build and run commands if its validated successfully, error otherwise.
func ValidateAndGetPushDevfileCommands(data data.DevfileData, devfileInitCmd, devfileBuildCmd, devfileRunCmd string) (commandMap PushCommandsMap, err error) {
	var emptyCommand common.DevfileCommand
	commandMap = NewPushCommandMap()

	isInitCommandValid, isBuildCommandValid, isRunCommandValid := false, false, false

	initCommand, initCmdErr := GetInitCommand(data, devfileInitCmd)

	isInitCmdEmpty := reflect.DeepEqual(emptyCommand, initCommand)
	if isInitCmdEmpty && initCmdErr == nil {
		// If there was no init command specified through odo push and no default init command in the devfile, default validate to true since the init command is optional
		isInitCommandValid = true
		klog.V(3).Infof("No init command was provided")
	} else if !isInitCmdEmpty && initCmdErr == nil {
		isInitCommandValid = true
		commandMap[common.InitCommandGroupType] = initCommand
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
		commandMap[common.BuildCommandGroupType] = buildCommand
		klog.V(3).Infof("Build command: %v", buildCommand.Exec.Id)
	}

	runCommand, runCmdErr := GetRunCommand(data, devfileRunCmd)
	if runCmdErr == nil && !reflect.DeepEqual(emptyCommand, runCommand) {
		isRunCommandValid = true
		commandMap[common.RunCommandGroupType] = runCommand
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
		return commandMap, fmt.Errorf(commandErrors)
	}

	return commandMap, nil
}

// Need to update group on custom commands specified by odo flags
func updateGroupforCustomCommand(commandName string, groupType common.DevfileCommandGroupType, command common.DevfileCommand) common.DevfileCommand {
	// Update Group only for exec commands
	// Update Group only custom commands (specified by odo flags)
	// Update Group only when Group is not nil, devfile v2 might contain group for custom commands.
	if command.Exec != nil && commandName != "" && command.Exec.Group == nil {
		command.Exec.Group = &common.Group{Kind: groupType}
		return command
	}
	return command
}
