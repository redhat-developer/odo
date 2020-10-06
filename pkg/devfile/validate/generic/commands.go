package generic

import (
	"strings"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// ValidateExecCommand validates the given exec command must map to a valid container component
func ValidateExecCommand(command common.DevfileCommand, components []common.DevfileComponent) (err error) {

	if !command.IsExec() {
		return &InvalidCommandError{commandId: command.GetID(), commandType: "exec"}
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

// ValidateCompositeCommand checks that the specified composite command is valid
func ValidateCompositeCommand(command *common.DevfileCommand, parentCommands map[string]string, devfileCommands map[string]common.DevfileCommand, components []common.DevfileComponent) error {

	// Store the command ID in a map of parent commands
	parentCommands[command.Id] = command.Id

	if !command.IsComposite() {
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
			err := ValidateCompositeCommand(&subCommand, parentCommands, devfileCommands, components)
			if err != nil {
				// Don't wrap the error message here to make the error message more readable to the user
				return err
			}
		} else {
			err := ValidateExecCommand(subCommand, components)
			if err != nil {
				return &CompositeInvalidSubCommandError{commandId: command.Id, subCommandId: subCommand.GetID(), errorMsg: err.Error()}
			}
		}
	}
	return nil
}
