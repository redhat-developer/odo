package v2

import (
	"fmt"
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"reflect"
	"strings"
)

// GetCommands returns the slice of Command objects parsed from the Devfile
func (d *DevfileV2) GetCommands(options common.DevfileOptions) ([]v1.Command, error) {

	if reflect.DeepEqual(options, common.DevfileOptions{}) {
		return d.Commands, nil
	}

	var commands []v1.Command
	for _, command := range d.Commands {
		// Filter Command Attributes
		filterIn, err := common.FilterDevfileObject(command.Attributes, options)
		if err != nil {
			return nil, err
		} else if !filterIn {
			continue
		}

		// Filter Command Type - Exec, Composite, etc.
		commandType, err := common.GetCommandType(command)
		if err != nil {
			return nil, err
		}
		if options.CommandOptions.CommandType != "" && commandType != options.CommandOptions.CommandType {
			continue
		}

		// Filter Command Group Kind - Run, Build, etc.
		commandGroup := common.GetGroup(command)
		// exclude conditions:
		// 1. options group is present and command group is present but does not match
		// 2. options group is present and command group is not present
		if options.CommandOptions.CommandGroupKind != "" && ((commandGroup != nil && options.CommandOptions.CommandGroupKind != commandGroup.Kind) || commandGroup == nil) {
			continue
		}

		if options.FilterByName == "" || command.Id == options.FilterByName {
			commands = append(commands, command)
		}
	}

	return commands, nil
}

// AddCommands adds the slice of Command objects to the Devfile's commands
// a command is considered as invalid if it is already defined
// command list passed in will be all processed, and returns a total error of all invalid commands
func (d *DevfileV2) AddCommands(commands []v1.Command) error {
	var errorsList []string
	for _, command := range commands {
		var err error
		for _, devfileCommand := range d.Commands {
			if command.Id == devfileCommand.Id {
				err = &common.FieldAlreadyExistError{Name: command.Id, Field: "command"}
				errorsList = append(errorsList, err.Error())
				break
			}
		}
		if err == nil {
			d.Commands = append(d.Commands, command)
		}
	}
	if len(errorsList) > 0 {
		return fmt.Errorf("errors while adding commands:\n%s", strings.Join(errorsList, "\n"))
	}
	return nil
}

// UpdateCommand updates the command with the given id
// return an error if the command is not found
func (d *DevfileV2) UpdateCommand(command v1.Command) error {
	for i := range d.Commands {
		if d.Commands[i].Id == command.Id {
			d.Commands[i] = command
			return nil
		}
	}
	return fmt.Errorf("update command failed: command %s not found", command.Id)
}

// DeleteCommand removes the specified command
func (d *DevfileV2) DeleteCommand(id string) error {

	for i := range d.Commands {
		if d.Commands[i].Id == id {
			d.Commands = append(d.Commands[:i], d.Commands[i+1:]...)
			return nil
		}
	}

	return &common.FieldNotFoundError{
		Field: "command",
		Name:  id,
	}
}
