package validate

import (
	"fmt"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	genericValidation "github.com/openshift/odo/pkg/devfile/validate/generic"
)

// validateCommands validates all the devfile commands. If there are commands with duplicate IDs, an error is returned
func validateCommands(commands []common.DevfileCommand, commandsMap map[string]common.DevfileCommand, components []common.DevfileComponent) (err error) {
	processedCommands := make(map[string]string, len(commands))

	for _, command := range commands {
		// Check if the command is in the list of already processed commands
		// If there's a hit, it means more than one command share the same ID and we should error out
		commandID := command.SetIDToLower()
		if _, exists := processedCommands[commandID]; exists {
			return fmt.Errorf("devfile has duplicate command IDs %q", commandID)
		}
		processedCommands[commandID] = commandID

		err = validateCommand(command, commandsMap, components)
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

	// devfile command type for odo must be exec or composite
	if command.Exec == nil && command.Composite == nil {
		return &UnsupportedOdoCommandError{commandId: command.GetID()}
	}

	// If the command is a composite command, need to validate that it is valid
	if command.IsComposite() {
		parentCommands := make(map[string]string)
		return validateCompositeCommand(&command, parentCommands, devfileCommands, components)
	}

	err = validateExecCommand(command, components)

	return err
}

func validateExecCommand(command common.DevfileCommand, components []common.DevfileComponent) error {

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

	err := genericValidation.ValidateExecCommand(command, components)

	return err
}

// validateCompositeCommand checks that the specified composite command is valid
func validateCompositeCommand(compositeCommand *common.DevfileCommand, parentCommands map[string]string, devfileCommands map[string]common.DevfileCommand, components []common.DevfileComponent) error {
	if compositeCommand.Composite.Group != nil && compositeCommand.Composite.Group.Kind == common.RunCommandGroupType {
		return &CompositeRunKindError{}
	}

	err := genericValidation.ValidateCompositeCommand(compositeCommand, parentCommands, devfileCommands, components)

	return err
}
