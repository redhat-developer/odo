package validation

import (
	"fmt"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// ValidateCommands validates the devfile commands and checks:
// 1. there are no duplicate command ids
// 2. the command type is not invalid
// 3. if a command is part of a command group, there is a single default command
func ValidateCommands(commands []v1alpha2.Command, components []v1alpha2.Component) (err error) {
	groupKindCommandMap := make(map[v1alpha2.CommandGroupKind][]v1alpha2.Command)
	commandMap := getCommandsMap(commands)

	err = v1alpha2.CheckDuplicateKeys(commands)
	if err != nil {
		return err
	}

	for _, command := range commands {
		// parentCommands is a map to keep a track of all the parent commands when validating the composite command's subcommands recursively
		parentCommands := make(map[string]string)
		err = validateCommand(command, parentCommands, commandMap, components)
		if err != nil {
			return resolveErrorMessageWithImportAttributes(err, command.Attributes)
		}

		commandGroup := getGroup(command)
		if commandGroup != nil {
			groupKindCommandMap[commandGroup.Kind] = append(groupKindCommandMap[commandGroup.Kind], command)
		}
	}

	var groupErrorsList []string
	for groupKind, commands := range groupKindCommandMap {
		if err = validateGroup(commands); err != nil {
			groupErrorsList = append(groupErrorsList, fmt.Sprintf("command group %s error - %s", groupKind, err.Error()))
		}
	}

	if len(groupErrorsList) > 0 {
		groupErrors := strings.Join(groupErrorsList, "\n")
		err = fmt.Errorf("\n%s", groupErrors)
	}

	return err
}

// validateCommand validates a given devfile command where parentCommands is a map to track all the parent commands when validating
// the composite command's subcommands recursively and devfileCommands is a map of command id to the devfile command
func validateCommand(command v1alpha2.Command, parentCommands map[string]string, devfileCommands map[string]v1alpha2.Command, components []v1alpha2.Component) error {

	switch {
	case command.Composite != nil:
		return validateCompositeCommand(&command, parentCommands, devfileCommands, components)
	case command.Exec != nil || command.Apply != nil:
		return validateCommandComponent(command, components)
	default:
		return &InvalidCommandTypeError{commandId: command.Id}
	}

}

// validateGroup validates commands belonging to a specific group kind. If there are multiple commands belonging to the same group:
// 1. without any default, err out
// 2. with more than one default, err out
func validateGroup(commands []v1alpha2.Command) error {
	defaultCommandCount := 0
	var defaultCommands []v1alpha2.Command
	if len(commands) > 1 {
		for _, command := range commands {
			if getGroup(command).IsDefault {
				defaultCommandCount++
				defaultCommands = append(defaultCommands, command)
			}
		}
	} else {
		return nil
	}

	if defaultCommandCount == 0 {
		return fmt.Errorf("there should be exactly one default command, currently there is no default command")
	} else if defaultCommandCount > 1 {
		var commandsReferenceList []string
		for _, command := range defaultCommands {
			commandsReferenceList = append(commandsReferenceList,
				resolveErrorMessageWithImportAttributes(fmt.Errorf("command: %s", command.Id), command.Attributes).Error())
		}
		commandsReference := strings.Join(commandsReferenceList, "; ")
		// example: there should be exactly one default command, currently there is more than one default command;
		// command: <id1>; command: <id2>, imported from uri: http://127.0.0.1:8080, in parent overrides from main devfile"
		return fmt.Errorf("there should be exactly one default command, currently there is more than one default command; %s", commandsReference)
	}

	return nil
}

// getGroup returns the group the command belongs to, or nil if the command does not belong to a group
func getGroup(command v1alpha2.Command) *v1alpha2.CommandGroup {
	switch {
	case command.Composite != nil:
		return command.Composite.Group
	case command.Exec != nil:
		return command.Exec.Group
	case command.Apply != nil:
		return command.Apply.Group
	case command.Custom != nil:
		return command.Custom.Group

	default:
		return nil
	}
}

// validateCommandComponent validates the given exec or apply command, the command should map to a valid container component
func validateCommandComponent(command v1alpha2.Command, components []v1alpha2.Component) error {

	if command.Exec == nil && command.Apply == nil {
		return &InvalidCommandError{commandId: command.Id, reason: "should be of type exec or apply"}
	}

	var commandComponent string
	if command.Exec != nil {
		commandComponent = command.Exec.Component
	} else if command.Apply != nil {
		commandComponent = command.Apply.Component
	}

	// must map to a container component
	for _, component := range components {
		if component.Container != nil && commandComponent == component.Name {
			return nil
		}
	}
	return &InvalidCommandError{commandId: command.Id, reason: "command does not map to a container component"}
}

// validateCompositeCommand checks that the specified composite command is valid. The command:
// 1. should not reference itself via a subcommand
// 2. should not indirectly reference itself via a subcommand which is a composite command
// 3. should reference a valid devfile command
// 4. should have a valid exec sub command
// where parentCommands is a map to track all the parent commands when validating the composite command's subcommands recursilvely
// and devfileCommands is a map of command id to the devfile command
func validateCompositeCommand(command *v1alpha2.Command, parentCommands map[string]string, devfileCommands map[string]v1alpha2.Command, components []v1alpha2.Component) error {

	// Store the command ID in a map of parent commands
	parentCommands[command.Id] = command.Id

	if command.Composite == nil {
		return &InvalidCommandError{commandId: command.Id, reason: "should be of type composite"}
	}

	// Loop over the commands and validate that each command points to a command that's in the devfile
	for _, cmd := range command.Composite.Commands {
		if strings.ToLower(cmd) == command.Id {
			return &InvalidCommandError{commandId: command.Id, reason: "composite command cannot reference itself"}
		}

		// Don't allow commands to indirectly reference themselves, so check if the command equals any of the parent commands in the command tree
		_, ok := parentCommands[strings.ToLower(cmd)]
		if ok {
			return &InvalidCommandError{commandId: command.Id, reason: "composite command cannot indirectly reference itself"}
		}

		subCommand, ok := devfileCommands[strings.ToLower(cmd)]
		if !ok {
			return &InvalidCommandError{commandId: command.Id, reason: fmt.Sprintf("the command %q mentioned in the composite command does not exist in the devfile", cmd)}
		}

		err := validateCommand(subCommand, parentCommands, devfileCommands, components)
		if err != nil {
			return err
		}
		delete(parentCommands, cmd)
	}
	return nil
}
