package libdevfile

import (
	"fmt"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/util"
)

// parallelCompositeCommand is a command implementation that represents parallel composite commands
type parallelCompositeCommand struct {
	command    v1alpha2.Command
	devfileObj parser.DevfileObj
}

// newParallelCompositeCommand creates a new command implementation which will execute the provided commands in parallel
func newParallelCompositeCommand(devfileObj parser.DevfileObj, command v1alpha2.Command) *parallelCompositeCommand {
	return &parallelCompositeCommand{
		command:    command,
		devfileObj: devfileObj,
	}
}

func (o *parallelCompositeCommand) CheckValidity() error {
	allCommands, err := allCommandsMap(o.devfileObj)
	if err != nil {
		return err
	}
	cmds := o.command.Composite.Commands
	for _, cmd := range cmds {
		if _, ok := allCommands[strings.ToLower(cmd)]; !ok {
			return fmt.Errorf("composite command %q has command %v not found in devfile", cmd, o.command.Id)
		}
	}
	return nil
}

// Execute loops over each command and executes them in parallel
func (o *parallelCompositeCommand) Execute(handler Handler) error {
	allCommands, err := allCommandsMap(o.devfileObj)
	if err != nil {
		return err
	}
	commandExecs := util.NewConcurrentTasks(len(o.command.Composite.Commands))
	for _, devfileCmd := range o.command.Composite.Commands {
		cmd, err2 := newCommand(o.devfileObj, allCommands[devfileCmd])
		if err2 != nil {
			return err2
		}
		commandExecs.Add(util.ConcurrentTask{
			ToRun: func(errChannel chan error) {
				err3 := cmd.Execute(handler)
				if err3 != nil {
					errChannel <- err3
				}
			},
		})
	}
	err = commandExecs.Run()
	if err != nil {
		return fmt.Errorf("parallel command execution failed: %w", err)
	}
	return nil
}
