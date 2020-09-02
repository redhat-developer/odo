package validate

import (
	"fmt"
	"strings"

	adapterCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/pkg/errors"
)

// validateCommands validates all the devfile commands
func validateCommands(commands []common.DevfileCommand, components []common.DevfileComponent) (err error) {

	commandsMap := adapterCommon.GetCommandsMap(commands)

	for _, command := range commands {
		err = validateCommand(command, commandsMap, components)
		if err != nil {
			return errors.Wrap(err, "test")
			// return fmt.Errorf("command %s has validation error: %v", command.GetID(), err)
		}
	}

	return /* fmt.Errorf("should err out") */
}

func validateCommand(command common.DevfileCommand, devfileCommands map[string]common.DevfileCommand, components []common.DevfileComponent) (err error) {

	// type must be exec or composite
	if command.Exec == nil && command.Composite == nil {
		return fmt.Errorf("command must be of type \"exec\" or \"composite\"")
	}

	// If the command is a composite command, need to validate that it is valid
	if command.Composite != nil {
		parentCommands := make(map[string]string)
		// commandsMap := adapterCommon.GetCommandsMap(commands)
		return validateCompositeCommand(command.Composite, parentCommands, devfileCommands, components)
	}

	// component must be specified
	if command.Exec.Component == "" {
		return fmt.Errorf("exec commands must reference a component")
	}

	// must specify a command
	if command.Exec.CommandLine == "" {
		return fmt.Errorf("exec commands must have a command")
	}

	// must map to a container component
	// components := GetDevfileContainerComponents(data)

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

// validateCompositeCommand checks that the specified composite command is valid
func validateCompositeCommand(compositeCommand *common.Composite, parentCommands map[string]string, devfileCommands map[string]common.DevfileCommand, components []common.DevfileComponent) error {
	if compositeCommand.Group != nil && compositeCommand.Group.Kind == common.RunCommandGroupType {
		return fmt.Errorf("composite commands of run Kind are not supported currently")
	}

	// Store the command ID in a map of parent commands
	parentCommands[compositeCommand.Id] = compositeCommand.Id

	// Loop over the commands and validate that each command points to a command that's in the devfile
	for _, command := range compositeCommand.Commands {
		if strings.ToLower(command) == compositeCommand.Id {
			return fmt.Errorf("the composite command %q cannot reference itself", compositeCommand.Id)
		}

		// Don't allow commands to indirectly reference themselves, so check if the command equals any of the parent commands in the command tree
		_, ok := parentCommands[strings.ToLower(command)]
		if ok {
			return fmt.Errorf("the composite command %q cannot indirectly reference itself", compositeCommand.Id)
		}

		subCommand, ok := devfileCommands[strings.ToLower(command)]
		if !ok {
			return fmt.Errorf("the command %q mentioned in the composite command %q does not exist in the devfile", command, compositeCommand.Id)
		}

		if subCommand.Composite != nil {
			// Recursively validate the composite subcommand
			err := validateCompositeCommand(subCommand.Composite, parentCommands, devfileCommands, components)
			if err != nil {
				// Don't wrap the error message here to make the error message more readable to the user
				return err
			}
		} else {
			err := validateCommand(subCommand, devfileCommands, components)
			if err != nil {
				return errors.Wrapf(err, "the composite command %q references an invalid command %q", compositeCommand.Id, subCommand.GetID())
			}
		}
	}
	return nil
}
