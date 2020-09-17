package common

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/openshift/odo/pkg/devfile/parser/data"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

// command encapsulates a command meant to be executed either directly or as part of a composite
type command interface {
	Execute(show bool) error
}

// New returns a new command implementation based on the specified devfile command and the known commands
func New(devfile common.DevfileCommand, knowCommands map[string]common.DevfileCommand, executor commandExecutor) (command, error) {
	composite := devfile.Composite
	if composite != nil {
		cmds := composite.Commands
		components := make([]command, 0, len(cmds))
		for _, cmd := range cmds {
			if devfileCommand, ok := knowCommands[strings.ToLower(cmd)]; ok {
				c, err := New(devfileCommand, knowCommands, executor)
				if err != nil {
					return nil, errors.Wrapf(err, "couldn't create command %s", cmd)
				}
				components = append(components, c)
			} else {
				return nil, fmt.Errorf("composite command %q has command %v not found in devfile", cmd, devfile)
			}
		}
		if composite.Parallel {
			return newParallelCompositeCommand(components...), nil
		}
		return newCompositeCommand(components...), nil
	} else {
		return newSimpleCommand(devfile, executor)
	}
}

// getCommand iterates through the devfile commands and returns the devfile command associated with the group
// commands mentioned via the flags are passed via commandName, empty otherwise
func getCommand(data data.DevfileData, commandName string, groupType common.DevfileCommandGroupType) (supportedCommand common.DevfileCommand, err error) {

	var command common.DevfileCommand

	if commandName == "" {
		command, err = getCommandFromDevfile(data, groupType)
	} else if commandName != "" {
		command, err = getCommandFromFlag(data, groupType, commandName)
	}

	return command, err
}

// getCommandFromDevfile iterates through the devfile commands and returns the command associated with the group
func getCommandFromDevfile(data data.DevfileData, groupType common.DevfileCommandGroupType) (supportedCommand common.DevfileCommand, err error) {
	commands := data.GetCommands()
	var onlyCommand common.DevfileCommand

	// validate the command groups before searching for a command match
	// if the command groups are invalid, err out
	// we only validate when the push command flags are absent,
	// since we know the command kind from the push flags
	err = validateCommandsForGroup(data, groupType)
	if err != nil {
		return common.DevfileCommand{}, err
	}

	for _, command := range commands {
		cmdGroup := command.GetGroup()
		if cmdGroup != nil && cmdGroup.Kind == groupType {
			if cmdGroup.IsDefault {
				return command, nil
			} else if reflect.DeepEqual(onlyCommand, common.DevfileCommand{}) {
				// return the only remaining command for the group if there is no default command
				// NOTE: we return outside the for loop since the next iteration can have a default command
				onlyCommand = command
			}
		}
	}

	// if default command is not found return the first command found for the matching type.
	if !reflect.DeepEqual(onlyCommand, common.DevfileCommand{}) {
		return onlyCommand, nil
	}

	msg := fmt.Sprintf("the command group of kind \"%v\" is not found in the devfile", groupType)
	// if run command or test command is not found in devfile then it is an error
	if groupType == common.RunCommandGroupType || groupType == common.TestCommandGroupType {
		err = fmt.Errorf(msg)
	} else {
		klog.V(2).Info(msg)
	}

	return
}

// getCommandFromFlag iterates through the devfile commands and returns the command specified associated with the group
func getCommandFromFlag(data data.DevfileData, groupType common.DevfileCommandGroupType, commandName string) (supportedCommand common.DevfileCommand, err error) {
	commands := data.GetCommands()

	for _, command := range commands {
		if command.GetID() == commandName {

			// Update Group only custom commands (specified by odo flags)
			command = updateCommandGroupIfReqd(groupType, command)

			// we have found the command with name, its groupType Should match to the flag
			// e.g --build-command "mybuild"
			// exec:
			//   id: mybuild
			//   group:
			//     kind: build
			cmdGroup := command.GetGroup()
			if cmdGroup != nil && cmdGroup.Kind != groupType {
				return command, fmt.Errorf("command group mismatched, command %s is of group %v in devfile.yaml", commandName, command.Exec.Group.Kind)
			}

			return command, nil
		}
	}

	// if any command specified via flag is not found in devfile then it is an error.
	err = fmt.Errorf("the command \"%v\" is not found in the devfile", commandName)

	return
}

// validateCommandsForGroup validates the commands in a devfile for a group
// 1. multiple commands belonging to a single group should have at least one default
// 2. multiple commands belonging to a single group cannot have more than one default
func validateCommandsForGroup(data data.DevfileData, groupType common.DevfileCommandGroupType) error {

	commands := getCommandsByGroup(data, groupType)

	defaultCommandCount := 0

	if len(commands) > 1 {
		for _, command := range commands {
			if command.GetGroup().IsDefault {
				defaultCommandCount++
			}
		}
	} else {
		// if there is only one command, it is the default command for the group
		defaultCommandCount = 1
	}

	if defaultCommandCount == 0 {
		return fmt.Errorf("there should be exactly one default command for command group %v, currently there is no default command", groupType)
	} else if defaultCommandCount > 1 {
		return fmt.Errorf("there should be exactly one default command for command group %v, currently there is more than one default command", groupType)
	}

	return nil
}

// GetBuildCommand iterates through the components in the devfile and returns the build command
func GetBuildCommand(data data.DevfileData, devfileBuildCmd string) (buildCommand common.DevfileCommand, err error) {

	return getCommand(data, devfileBuildCmd, common.BuildCommandGroupType)
}

// GetDebugCommand iterates through the components in the devfile and returns the debug command
func GetDebugCommand(data data.DevfileData, devfileDebugCmd string) (debugCommand common.DevfileCommand, err error) {
	return getCommand(data, devfileDebugCmd, common.DebugCommandGroupType)
}

// GetRunCommand iterates through the components in the devfile and returns the run command
func GetRunCommand(data data.DevfileData, devfileRunCmd string) (runCommand common.DevfileCommand, err error) {

	return getCommand(data, devfileRunCmd, common.RunCommandGroupType)
}

// GetTestCommand iterates through the components in the devfile and returns the test command
func GetTestCommand(data data.DevfileData, devfileTestCmd string) (runCommand common.DevfileCommand, err error) {

	return getCommand(data, devfileTestCmd, common.TestCommandGroupType)
}

// ValidateAndGetPushDevfileCommands validates the build and the run command,
// if provided through odo push or else checks the devfile for devBuild and devRun.
// It returns the build and run commands if its validated successfully, error otherwise.
func ValidateAndGetPushDevfileCommands(data data.DevfileData, devfileBuildCmd, devfileRunCmd string) (commandMap PushCommandsMap, err error) {
	var emptyCommand common.DevfileCommand
	commandMap = NewPushCommandMap()

	isBuildCommandValid, isRunCommandValid := false, false

	buildCommand, buildCmdErr := GetBuildCommand(data, devfileBuildCmd)

	isBuildCmdEmpty := reflect.DeepEqual(emptyCommand, buildCommand)
	if isBuildCmdEmpty && buildCmdErr == nil {
		// If there was no build command specified through odo push and no default build command in the devfile, default validate to true since the build command is optional
		isBuildCommandValid = true
		klog.V(2).Infof("No build command was provided")
	} else if !reflect.DeepEqual(emptyCommand, buildCommand) && buildCmdErr == nil {
		isBuildCommandValid = true
		commandMap[common.BuildCommandGroupType] = buildCommand
		klog.V(2).Infof("Build command: %v", buildCommand.GetID())
	}

	runCommand, runCmdErr := GetRunCommand(data, devfileRunCmd)
	if runCmdErr == nil && !reflect.DeepEqual(emptyCommand, runCommand) {
		isRunCommandValid = true
		commandMap[common.RunCommandGroupType] = runCommand
		klog.V(2).Infof("Run command: %v", runCommand.GetID())
	}

	// If either command had a problem, return an empty list of commands and an error
	if !isBuildCommandValid || !isRunCommandValid {
		commandErrors := ""

		if buildCmdErr != nil {
			commandErrors += fmt.Sprintf("\n%s", buildCmdErr.Error())
		}
		if runCmdErr != nil {
			commandErrors += fmt.Sprintf("\n%s", runCmdErr.Error())
		}
		return commandMap, fmt.Errorf(commandErrors)
	}

	return commandMap, nil
}

// Need to update group on custom commands specified by odo flags
func updateCommandGroupIfReqd(groupType common.DevfileCommandGroupType, command common.DevfileCommand) common.DevfileCommand {
	// Update Group only for exec commands
	// Update Group only when Group is not nil, devfile v2 might contain group for custom commands.
	if command.Exec != nil && command.Exec.Group == nil {
		command.Exec.Group = &common.Group{Kind: groupType}
		return command
	}
	return command
}

// ValidateAndGetDebugDevfileCommands validates the debug command
func ValidateAndGetDebugDevfileCommands(data data.DevfileData, devfileDebugCmd string) (pushDebugCommand common.DevfileCommand, err error) {
	var emptyCommand common.DevfileCommand

	isDebugCommandValid := false
	debugCommand, debugCmdErr := GetDebugCommand(data, devfileDebugCmd)
	if debugCmdErr == nil && !reflect.DeepEqual(emptyCommand, debugCommand) {
		isDebugCommandValid = true
		klog.V(2).Infof("Debug command: %v", debugCommand.Id)
	}

	if !isDebugCommandValid {
		commandErrors := ""
		if debugCmdErr != nil {
			commandErrors += debugCmdErr.Error()
		}
		return common.DevfileCommand{}, fmt.Errorf(commandErrors)
	}

	return debugCommand, nil
}

// ValidateAndGetTestDevfileCommands validates the test command
func ValidateAndGetTestDevfileCommands(data data.DevfileData, devfileTestCmd string) (testCommand common.DevfileCommand, err error) {
	var emptyCommand common.DevfileCommand
	isTestCommandValid := false
	testCommand, testCmdErr := GetTestCommand(data, devfileTestCmd)
	if testCmdErr == nil && !reflect.DeepEqual(emptyCommand, testCommand) {
		isTestCommandValid = true
		klog.V(2).Infof("Test command: %v", testCommand.GetID())
	}

	if !isTestCommandValid && testCmdErr != nil {
		return common.DevfileCommand{}, testCmdErr
	}

	return testCommand, nil
}
