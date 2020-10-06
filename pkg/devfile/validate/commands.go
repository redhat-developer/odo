package validate

import (
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	genericValidation "github.com/openshift/odo/pkg/devfile/validate/generic"
)

// validateCommands validates the devfile commands:
// 1. checks if its either an exec or composite command
// 2. checks if the composite is a non run kind command
// 3. checks if the command is a valid exec or composite command
func validateCommands(commands []common.DevfileCommand, commandsMap map[string]common.DevfileCommand, components []common.DevfileComponent) (err error) {

	for _, command := range commands {
		err = validateCommand(command, commandsMap, components)
		if err != nil {
			return err
		}
	}

	err = genericValidation.ValidateCommands(commands, commandsMap, components)

	return err
}

// validateCommand validates the given command
// 1. command has to be of type exec or composite,
// 2. if composite command, it should not be of kind run
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

	return
}

// validateCompositeCommand checks that the specified composite command is valid in odo ie; it should not be of kind run
func validateCompositeCommand(compositeCommand *common.DevfileCommand, parentCommands map[string]string, devfileCommands map[string]common.DevfileCommand, components []common.DevfileComponent) error {
	if compositeCommand.Composite.Group != nil && compositeCommand.Composite.Group.Kind == common.RunCommandGroupType {
		return &CompositeRunKindError{}
	}

	return nil
}
