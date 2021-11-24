package common

import (
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// applyCommand is a command implementation for Apply commands
type applyCommand struct {
	adapter   commandExecutor
	id        string
	component string
}

// newApplyCommand creates a new applyCommand instance, adapting the devfile-defined command to run in the target component's container
func newApplyCommand(command devfilev1.Command, executor commandExecutor) (command, error) {
	apply := command.Apply
	return &applyCommand{
		adapter:   executor,
		id:        command.Id,
		component: apply.Component,
	}, nil
}

func (s applyCommand) Execute(show bool) error {
	err := s.adapter.ApplyComponent(s.component)
	return err
}
