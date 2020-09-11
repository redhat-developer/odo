package validate

import (
	"strings"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// validateCommands validates all the devfile commands
func validateCommands(commands map[string]common.DevfileCommand, components []common.DevfileComponent) (err error) {

	for _, command := range commands {
		err = validateCommand(command, commands, components)
		if err != nil {
			return err
		}
	}

	return
}

// validateCommand validates the given command
// 1. command has to be of type exec or composite, if composite command is validated further
// 2. component should be present
// 3. commandline should be present
// 4. command must map to a valid container component
func validateCommand(command common.DevfileCommand, devfileCommands map[string]common.DevfileCommand, components []common.DevfileComponent) (err error) {

	// type must be exec or composite
	if command.Exec == nil && command.Composite == nil {
		return &UnsupportedOdoCommandError{commandId: command.GetID()}
	}

	// If the command is a composite command, need to validate that it is valid
	if command.Composite != nil {
		parentCommands := make(map[string]string)
		return validateCompositeCommand(&command, parentCommands, devfileCommands, components)
	}

	// component must be specified
	if command.Exec.Component == "" {
		return &ExecCommandMissingComponentError{commandId: command.GetID()}
	}

	// must specify a command
	if command.Exec.CommandLine == "" {
		return &ExecCommandMissingCommandLineError{commandId: command.GetID()}
	}

	// must map to a container component
	isComponentValid := false
	for _, component := range components {
		if component.IsContainer() && command.Exec.Component == component.Name {
			isComponentValid = true
		}
	}
	if !isComponentValid {
		return &ExecCommandInvalidContainerError{commandId: command.GetID()}
	}

	return
}

// validateCompositeCommand checks that the specified composite command is valid
func validateCompositeCommand(compositeCommand *common.DevfileCommand, parentCommands map[string]string, devfileCommands map[string]common.DevfileCommand, components []common.DevfileComponent) error {
	if compositeCommand.Composite.Group != nil && compositeCommand.Composite.Group.Kind == common.RunCommandGroupType {
		return &CompositeRunKindError{}
	}

	// Store the command ID in a map of parent commands
	parentCommands[compositeCommand.Id] = compositeCommand.Id

	// Loop over the commands and validate that each command points to a command that's in the devfile
	for _, command := range compositeCommand.Composite.Commands {
		if strings.ToLower(command) == compositeCommand.Id {
			return &CompositeDirectReferenceError{commandId: compositeCommand.Id}
		}

		// Don't allow commands to indirectly reference themselves, so check if the command equals any of the parent commands in the command tree
		_, ok := parentCommands[strings.ToLower(command)]
		if ok {
			return &CompositeIndirectReferenceError{commandId: compositeCommand.Id}
		}

		subCommand, ok := devfileCommands[strings.ToLower(command)]
		if !ok {
			return &CompositeMissingSubCommandError{commandId: compositeCommand.Id, subCommand: command}
		}

		if subCommand.Composite != nil {
			// Recursively validate the composite subcommand
			err := validateCompositeCommand(&subCommand, parentCommands, devfileCommands, components)
			if err != nil {
				// Don't wrap the error message here to make the error message more readable to the user
				return err
			}
		} else {
			err := validateCommand(subCommand, devfileCommands, components)
			if err != nil {
				return &CompositeInvalidSubCommandError{commandId: compositeCommand.Id, subCommandId: subCommand.GetID()}
			}
		}
	}
	return nil
}
