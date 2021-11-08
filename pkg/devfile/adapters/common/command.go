package common

import (
	"fmt"
	"reflect"
	"strings"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

// command encapsulates a command meant to be executed either directly or as part of a composite
type command interface {
	Execute(show bool) error
}

// New returns a new command implementation based on the specified devfile command and the known commands
func New(devfile devfilev1.Command, knowCommands map[string]devfilev1.Command, executor commandExecutor) (command, error) {
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
		if util.SafeGetBool(composite.Parallel) {
			return newParallelCompositeCommand(components...), nil
		}
		return newCompositeCommand(components...), nil
	} else if devfile.Exec != nil {
		return newExecCommand(devfile, executor)
	} else {
		return newApplyCommand(devfile, executor)
	}
}

// getCommand iterates through the devfile commands and returns the devfile command associated with the group
// commands mentioned via the flags are passed via commandName, empty otherwise
func getCommand(data data.DevfileData, commandName string, groupType devfilev1.CommandGroupKind) (supportedCommand devfilev1.Command, err error) {

	var command devfilev1.Command

	if commandName == "" {
		command, err = getCommandFromDevfile(data, groupType)
	} else if commandName != "" {
		command, err = getCommandFromFlag(data, groupType, commandName)
	}

	return command, err
}

// getCommandFromDevfile iterates through the devfile commands and returns the command associated with the group
func getCommandFromDevfile(data data.DevfileData, groupType devfilev1.CommandGroupKind) (supportedCommand devfilev1.Command, err error) {
	commands, err := data.GetCommands(parsercommon.DevfileOptions{})
	if err != nil {
		return devfilev1.Command{}, err
	}
	var onlyCommand devfilev1.Command

	for _, command := range commands {
		cmdGroup := parsercommon.GetGroup(command)
		if cmdGroup != nil && cmdGroup.Kind == groupType {
			if util.SafeGetBool(cmdGroup.IsDefault) {
				return command, nil
			} else if reflect.DeepEqual(onlyCommand, devfilev1.Command{}) {
				// return the only remaining command for the group if there is no default command
				// NOTE: we return outside the for loop since the next iteration can have a default command
				onlyCommand = command
			}
		}
	}

	// if default command is not found return the first command found for the matching type.
	if !reflect.DeepEqual(onlyCommand, devfilev1.Command{}) {
		return onlyCommand, nil
	}

	notFoundError := NoCommandForGroup{Group: groupType}
	// if run command or test command is not found in devfile then it is an error
	if groupType == devfilev1.RunCommandGroupKind || groupType == devfilev1.TestCommandGroupKind {
		err = notFoundError
	} else {
		klog.V(2).Info(notFoundError)
	}

	return
}

// getCommandFromFlag iterates through the devfile commands and returns the command specified associated with the group
func getCommandFromFlag(data data.DevfileData, groupType devfilev1.CommandGroupKind, commandName string) (supportedCommand devfilev1.Command, err error) {
	commands, err := data.GetCommands(parsercommon.DevfileOptions{})
	if err != nil {
		return devfilev1.Command{}, err
	}

	for _, command := range commands {
		if command.Id == commandName {

			// Update Group only custom commands (specified by odo flags)
			command = updateCommandGroupIfReqd(groupType, command)

			// we have found the command with name, its groupType Should match to the flag
			// e.g --build-command "mybuild"
			// exec:
			//   id: mybuild
			//   group:
			//     kind: build
			cmdGroup := parsercommon.GetGroup(command)
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

// GetBuildCommand iterates through the components in the devfile and returns the build command
func GetBuildCommand(data data.DevfileData, devfileBuildCmd string) (buildCommand devfilev1.Command, err error) {
	return getCommand(data, devfileBuildCmd, devfilev1.BuildCommandGroupKind)
}

// GetDebugCommand iterates through the components in the devfile and returns the debug command
func GetDebugCommand(data data.DevfileData, devfileDebugCmd string) (debugCommand devfilev1.Command, err error) {
	return getCommand(data, devfileDebugCmd, devfilev1.DebugCommandGroupKind)
}

// GetRunCommand iterates through the components in the devfile and returns the run command
func GetRunCommand(data data.DevfileData, devfileRunCmd string) (runCommand devfilev1.Command, err error) {
	return getCommand(data, devfileRunCmd, devfilev1.RunCommandGroupKind)
}

// GetTestCommand iterates through the components in the devfile and returns the test command
func GetTestCommand(data data.DevfileData, devfileTestCmd string) (runCommand devfilev1.Command, err error) {
	return getCommand(data, devfileTestCmd, devfilev1.TestCommandGroupKind)
}

// ValidateAndGetPushDevfileCommands validates the build and the run command,
// if provided through odo push or else checks the devfile for devBuild and devRun.
// It returns the build and run commands if its validated successfully, error otherwise.
func ValidateAndGetPushDevfileCommands(data data.DevfileData, devfileBuildCmd, devfileRunCmd string) (commandMap PushCommandsMap, err error) {
	var emptyCommand devfilev1.Command
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
		commandMap[devfilev1.BuildCommandGroupKind] = buildCommand
		klog.V(2).Infof("Build command: %v", buildCommand.Id)
	}

	runCommand, runCmdErr := GetRunCommand(data, devfileRunCmd)
	if runCmdErr == nil && !reflect.DeepEqual(emptyCommand, runCommand) {
		isRunCommandValid = true
		commandMap[devfilev1.RunCommandGroupKind] = runCommand
		klog.V(2).Infof("Run command: %v", runCommand.Id)
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
func updateCommandGroupIfReqd(groupType devfilev1.CommandGroupKind, command devfilev1.Command) devfilev1.Command {
	// Update Group only for exec commands
	// Update Group only when Group is not nil, devfile v2 might contain group for custom commands.
	if command.Exec != nil && command.Exec.Group == nil {
		command.Exec.Group = &devfilev1.CommandGroup{Kind: groupType}
		return command
	}
	return command
}

// ValidateAndGetDebugDevfileCommands validates the debug command
func ValidateAndGetDebugDevfileCommands(data data.DevfileData, devfileDebugCmd string) (pushDebugCommand devfilev1.Command, err error) {
	var emptyCommand devfilev1.Command

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
		return devfilev1.Command{}, fmt.Errorf(commandErrors)
	}

	return debugCommand, nil
}

// ValidateAndGetTestDevfileCommands validates the test command
func ValidateAndGetTestDevfileCommands(data data.DevfileData, devfileTestCmd string) (testCommand devfilev1.Command, err error) {
	var emptyCommand devfilev1.Command
	isTestCommandValid := false
	testCommand, testCmdErr := GetTestCommand(data, devfileTestCmd)
	if testCmdErr == nil && !reflect.DeepEqual(emptyCommand, testCommand) {
		isTestCommandValid = true
		klog.V(2).Infof("Test command: %v", testCommand.Id)
	}

	if !isTestCommandValid && testCmdErr != nil {
		return devfilev1.Command{}, testCmdErr
	}

	return testCommand, nil
}
