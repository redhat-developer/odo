package libdevfile

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
)

// execCommand is a command implementation for exec commands
type execCommand struct {
	command    v1alpha2.Command
	devfileObj parser.DevfileObj
}

var _ command = (*execCommand)(nil)

// newExecCommand creates a new execCommand instance, adapting the devfile-defined command to run in the target component's
// container, modifying it to add environment variables or adapting the path as needed.
func newExecCommand(devfileObj parser.DevfileObj, command v1alpha2.Command) *execCommand {
	return &execCommand{
		command:    command,
		devfileObj: devfileObj,
	}
}

func (o *execCommand) CheckValidity() error {
	return nil
}

func (o *execCommand) Execute(handler Handler) error {
	return handler.Execute(o.command)
}
