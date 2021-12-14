package common

import "fmt"

// newCompositeCommand creates a new command implementation which will execute the provided commands in the specified order
func newCompositeCommand(cmds ...command) command {
	return compositeCommand{
		cmds: cmds,
	}
}

// compositeCommand is a command implementation that represents non-parallel composite commands
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

func (c compositeCommand) UnExecute() error {
	// UnExecute the commands in order
	for _, command := range c.cmds {
		err := command.UnExecute()
		if err != nil {
			return fmt.Errorf("command execution failed: %v", err)
		}
	}
	return nil
}
