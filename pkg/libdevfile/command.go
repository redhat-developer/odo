package libdevfile

import (
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"

	"github.com/redhat-developer/odo/pkg/util"
)

type command interface {
	CheckValidity() error
	Execute(handler Handler) error
}

// newCommand returns a command implementation, depending on the type of the command
func newCommand(devfileObj parser.DevfileObj, devfileCmd v1alpha2.Command) (command, error) {
	var cmd command

	commandType, err := common.GetCommandType(devfileCmd)
	if err != nil {
		return nil, err
	}

	switch commandType {

	case v1alpha2.ApplyCommandType:
		cmd = newApplyCommand(devfileObj, devfileCmd)

	case v1alpha2.CompositeCommandType:
		if util.SafeGetBool(devfileCmd.Composite.Parallel) {
			cmd = newParallelCompositeCommand(devfileObj, devfileCmd)
		}
		cmd = newCompositeCommand(devfileObj, devfileCmd)

	case v1alpha2.ExecCommandType:
		cmd = newExecCommand(devfileObj, devfileCmd)
	}

	if err = cmd.CheckValidity(); err != nil {
		return nil, err
	}
	return cmd, nil
}

// allCommandsMap returns a map of all commands in the devfile, indexed by Id
func allCommandsMap(devfileObj parser.DevfileObj) (map[string]v1alpha2.Command, error) {
	commands, err := devfileObj.Data.GetCommands(common.DevfileOptions{})
	if err != nil {
		return nil, err
	}

	commandMap := make(map[string]v1alpha2.Command, len(commands))
	for _, command := range commands {
		commandMap[strings.ToLower(command.Id)] = command
	}

	return commandMap, nil
}
