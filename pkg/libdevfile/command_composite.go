package libdevfile

import (
	"fmt"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
)

// compositeCommand is a command implementation that represents non-parallel composite commands
type compositeCommand struct {
	command    v1alpha2.Command
	devfileObj parser.DevfileObj
}

// newCompositeCommand creates a new command implementation which will execute the provided commands in the specified order
func newCompositeCommand(devfileObj parser.DevfileObj, command v1alpha2.Command) *compositeCommand {
	return &compositeCommand{
		command:    command,
		devfileObj: devfileObj,
	}
}

func (o *compositeCommand) CheckValidity() error {
	allCommands, err := allCommandsMap(o.devfileObj)
	if err != nil {
		return err
	}
	cmds := o.command.Composite.Commands
	for _, cmd := range cmds {
		if _, ok := allCommands[strings.ToLower(cmd)]; !ok {
			return fmt.Errorf("composite command %q references command %q not found in devfile", o.command.Id, cmd)
		}
	}
	return nil
}

// Execute loops over each command and executes them serially
func (o *compositeCommand) Execute(handler Handler) error {
	allCommands, err := allCommandsMap(o.devfileObj)
	if err != nil {
		return err
	}
	for _, devfileCmd := range o.command.Composite.Commands {
		cmd, err := newCommand(o.devfileObj, allCommands[strings.ToLower(devfileCmd)])
		if err != nil {
			return err
		}
		err = cmd.Execute(handler)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *compositeCommand) UnExecute() error {
	return nil
}
