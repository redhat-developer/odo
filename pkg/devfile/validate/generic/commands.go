package generic

import (
	"strings"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// validateCommands validates the devfile commands:
// 1. if there are commands with duplicate IDs, an error is returned
// 2. checks if its either a valid exec or composite command
func validateCommands(commands []common.DevfileCommand, commandsMap map[string]common.DevfileCommand, components []common.DevfileComponent) (err error) {
	processedCommands := make(map[string]string, len(commands))

	for _, command := range commands {
		// Check if the command is in the list of already processed commands
		// If there's a hit, it means more than one command share the same ID and we should error out
		commandID := command.SetIDToLower()
		if _, exists := processedCommands[commandID]; exists {
			return &DuplicateCommandError{commandId: commandID}
		}
		processedCommands[commandID] = commandID

		err = validateCommand(command, commandsMap, components)
		if err != nil {
			return err
		}
	}

	return
}

// validateCommand validates a given devfile command
func validateCommand(command common.DevfileCommand, devfileCommands map[string]common.DevfileCommand, components []common.DevfileComponent) (err error) {

	// If the command is a composite command, need to validate that it is valid
	if command.Composite != nil {
		parentCommands := make(map[string]string)
		return validateCompositeCommand(&command, parentCommands, devfileCommands, components)
	}

	return validateExecCommand(command, components)
}

// validateExecCommand validates the given exec command, the command should:
// 1. have a component
// 2. have a command line
// 3. map to a valid container component
func validateExecCommand(command common.DevfileCommand, components []common.DevfileComponent) (err error) {

	if command.Exec == nil {
		return &InvalidCommandError{commandId: command.GetID(), commandType: "exec"}
	}

	// TODO - Remove component and command line check when devfile spec is finalized for 2.0.0
	// since these are required fields in a devfile.yaml

	// component must be specified
	if command.GetExecComponent() == "" {
		return &ExecCommandMissingComponentError{commandId: command.GetID()}
	}

	// must specify a command
	if command.GetExecCommandLine() == "" {
		return &ExecCommandMissingCommandLineError{commandId: command.GetID()}
	}

	// must map to a container component
	isComponentValid := false
	for _, component := range components {
		if component.Container != nil && command.Exec.Component == component.Name {
			isComponentValid = true
		}
	}
	if !isComponentValid {
		return &ExecCommandInvalidContainerError{commandId: command.GetID()}
	}

	return
}

// validateCompositeCommand checks that the specified composite command is valid. The command:
// 1. should not reference itself via s subcommand
// 2. should not indirectly reference itself via a subcommand which is a composite command
// 3. should reference a valid devfile command
// 4. should have a valid exec sub command
func validateCompositeCommand(command *common.DevfileCommand, parentCommands map[string]string, devfileCommands map[string]common.DevfileCommand, components []common.DevfileComponent) error {

	// Store the command ID in a map of parent commands
	parentCommands[command.Id] = command.Id

	if command.Composite == nil {
		return &InvalidCommandError{commandId: command.GetID(), commandType: "composite"}
	}

	// Loop over the commands and validate that each command points to a command that's in the devfile
	for _, cmd := range command.Composite.Commands {
		if strings.ToLower(cmd) == command.Id {
			return &CompositeDirectReferenceError{commandId: command.Id}
		}

		// Don't allow commands to indirectly reference themselves, so check if the command equals any of the parent commands in the command tree
		_, ok := parentCommands[strings.ToLower(cmd)]
		if ok {
			return &CompositeIndirectReferenceError{commandId: command.Id}
		}

		subCommand, ok := devfileCommands[strings.ToLower(cmd)]
		if !ok {
			return &CompositeMissingSubCommandError{commandId: command.Id, subCommand: cmd}
		}

		if subCommand.Composite != nil {
			// Recursively validate the composite subcommand
			err := validateCompositeCommand(&subCommand, parentCommands, devfileCommands, components)
			if err != nil {
				// Don't wrap the error message here to make the error message more readable to the user
				return err
			}
		} else {
			err := validateExecCommand(subCommand, components)
			if err != nil {
				return &CompositeInvalidSubCommandError{commandId: command.Id, subCommandId: subCommand.GetID(), errorMsg: err.Error()}
			}
		}
	}
	return nil
}
