package libdevfile

import (
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

// applyCommand is a command implementation for Apply commands
type applyCommand struct {
	command    v1alpha2.Command
	devfileObj parser.DevfileObj
}

// newApplyCommand creates a new applyCommand instance
func newApplyCommand(devfileObj parser.DevfileObj, command v1alpha2.Command) *applyCommand {
	return &applyCommand{
		command:    command,
		devfileObj: devfileObj,
	}
}

func (o *applyCommand) CheckValidity() error {
	return nil
}

func (o *applyCommand) Execute(handler Handler) error {
	devfileComponents, err := o.devfileObj.Data.GetComponents(common.DevfileOptions{
		FilterByName: o.command.Apply.Component,
	})
	if err != nil {
		return err
	}

	if len(devfileComponents) == 0 {
		return fmt.Errorf("component %q does not exists", o.command.Apply.Component)
	}

	if len(devfileComponents) != 1 {
		return fmt.Errorf("more than one component with the same name, should not happen")
	}

	component, err := newComponent(o.devfileObj, devfileComponents[0])
	if err != nil {
		return err
	}

	return component.Apply(handler)
}

func (o *applyCommand) UnExecute() error {
	return nil
}
