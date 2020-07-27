package common

import (
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
)

func newParallelCompositeCommand(cmds ...command) command {
	return parallelCompositeCommand{
		cmds: cmds,
	}
}

type parallelCompositeCommand struct {
	cmds []command
}

func (p parallelCompositeCommand) Execute(show bool) error {
	// Loop over each command and execute it in parallel
	commandExecs := util.NewConcurrentTasks(len(p.cmds))
	for _, command := range p.cmds {
		cmd := command // needed to prevent the lambda from capturing the value
		commandExecs.Add(util.ConcurrentTask{ToRun: func(errChannel chan error) {
			err := cmd.Execute(show)
			if err != nil {
				errChannel <- err
			}
		}})
	}

	err := commandExecs.Run()
	if err != nil {
		return errors.Wrap(err, "parallel command execution failed")
	}
	return nil
}
