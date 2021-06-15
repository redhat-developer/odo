package common

import (
	"fmt"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// GetGroup returns the group the command belongs to
func GetGroup(dc v1.Command) *v1.CommandGroup {
	switch {
	case dc.Composite != nil:
		return dc.Composite.Group
	case dc.Exec != nil:
		return dc.Exec.Group
	case dc.Apply != nil:
		return dc.Apply.Group
	case dc.Custom != nil:
		return dc.Custom.Group

	default:
		return nil
	}
}

// GetExecComponent returns the component of the exec command
func GetExecComponent(dc v1.Command) string {
	if dc.Exec != nil {
		return dc.Exec.Component
	}

	return ""
}

// GetExecCommandLine returns the command line of the exec command
func GetExecCommandLine(dc v1.Command) string {
	if dc.Exec != nil {
		return dc.Exec.CommandLine
	}

	return ""
}

// GetExecWorkingDir returns the working dir of the exec command
func GetExecWorkingDir(dc v1.Command) string {
	if dc.Exec != nil {
		return dc.Exec.WorkingDir
	}

	return ""
}

// GetApplyComponent returns the component of the apply command
func GetApplyComponent(dc v1.Command) string {
	if dc.Apply != nil {
		return dc.Apply.Component
	}

	return ""
}

// GetCommandType returns the command type of a given command
func GetCommandType(command v1.Command) (v1.CommandType, error) {
	switch {
	case command.Apply != nil:
		return v1.ApplyCommandType, nil
	case command.Composite != nil:
		return v1.CompositeCommandType, nil
	case command.Exec != nil:
		return v1.ExecCommandType, nil
	case command.Custom != nil:
		return v1.CustomCommandType, nil

	default:
		return "", fmt.Errorf("unknown command type")
	}
}

// GetCommandsMap returns a map of the command Id to the command
func GetCommandsMap(commands []v1.Command) map[string]v1.Command {
	commandMap := make(map[string]v1.Command, len(commands))
	for _, command := range commands {
		commandMap[command.Id] = command
	}
	return commandMap
}

// GetCommandsFromEvent returns the list of commands from the event name.
// If the event is a composite command, it returns the sub-commands from the tree
func GetCommandsFromEvent(commandsMap map[string]v1.Command, eventName string) []string {
	var commands []string

	if command, ok := commandsMap[eventName]; ok {
		if command.Composite != nil {
			for _, compositeSubCmd := range command.Composite.Commands {
				subCommands := GetCommandsFromEvent(commandsMap, compositeSubCmd)
				commands = append(commands, subCommands...)
			}
		} else {
			commands = append(commands, command.Id)
		}
	}

	return commands
}
