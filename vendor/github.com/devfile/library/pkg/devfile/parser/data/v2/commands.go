package v2

import (
	"strings"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

// GetCommands returns the slice of Command objects parsed from the Devfile
func (d *DevfileV2) GetCommands() map[string]v1.Command {

	commands := make(map[string]v1.Command, len(d.Commands))

	for _, command := range d.Commands {
		command.Id = strings.ToLower(command.Id)
		commands[command.Id] = command
	}

	return commands
}

// AddCommands adds the slice of Command objects to the Devfile's commands
// if a command is already defined, error out
func (d *DevfileV2) AddCommands(commands ...v1.Command) error {
	commandsMap := d.GetCommands()

	for _, command := range commands {
		if _, ok := commandsMap[command.Id]; !ok {
			d.Commands = append(d.Commands, command)
		} else {
			return &common.FieldAlreadyExistError{Name: command.Id, Field: "command"}
		}
	}
	return nil
}

// UpdateCommand updates the command with the given id
func (d *DevfileV2) UpdateCommand(command v1.Command) {
	for i := range d.Commands {
		if strings.ToLower(d.Commands[i].Id) == strings.ToLower(command.Id) {
			d.Commands[i] = command
			d.Commands[i].Id = strings.ToLower(d.Commands[i].Id)
		}
	}
}
