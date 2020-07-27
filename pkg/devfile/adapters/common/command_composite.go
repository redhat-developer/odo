package common

import "fmt"

func newCompositeCommand(cmds ...command) command {
	return compositeCommand{
		cmds: cmds,
	}
}

type compositeCommand struct {
	cmds []command
}

func (c compositeCommand) Execute(show bool) error {
	// Execute the commands in order
	for _, command := range c.cmds {
		err := command.Execute(show)
		if err != nil {
			return fmt.Errorf("command execution failed: %v", err)
		}
	}
	return nil
}
