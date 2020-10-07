package validate

import (
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// validateCommands validates the devfile commands:
// 1. checks if its either an exec or composite command
// 2. checks if the composite is a non run kind command
func validateCommands(commandsMap map[string]common.DevfileCommand) (err error) {

	for _, command := range commandsMap {
		err = validateCommand(command)
		if err != nil {
			return err
		}
	}

	return
}

// validateCommand validates the given command
// 1. command has to be of type exec or composite,
// 2. if composite command, it should not be of kind run
func validateCommand(command common.DevfileCommand) (err error) {

	// devfile command type for odo must be exec or composite
	if command.Exec == nil && command.Composite == nil {
		return &UnsupportedOdoCommandError{commandId: command.GetID()}
	}

	// If the command is a composite command, need to validate that it is valid
	if command.Composite != nil {
		return validateCompositeCommand(command)
	}

	return
}

// validateCompositeCommand checks that the specified composite command is valid in odo ie; it should not be of kind run
func validateCompositeCommand(command common.DevfileCommand) error {
	if command.Composite.Group != nil && command.Composite.Group.Kind == common.RunCommandGroupType {
		return &CompositeRunKindError{}
	}

	return nil
}
