package exports

import (
	"fmt"
	"syscall/js"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"

	"github.com/feloy/devfile-builder/wasm/pkg/global"
	"github.com/feloy/devfile-builder/wasm/pkg/utils"
)

func DeleteCommandWrapper(this js.Value, args []js.Value) interface{} {
	return result(
		deleteCommand(args[0].String()),
	)
}

func deleteCommand(name string) (map[string]interface{}, error) {

	err := checkCommandUsed(name)
	if err != nil {
		return nil, fmt.Errorf("error deleting command %q: %w", name, err)
	}
	err = global.Devfile.Data.DeleteCommand(name)
	if err != nil {
		return nil, err
	}
	return utils.GetContent()
}

func checkCommandUsed(name string) error {
	commands, err := global.Devfile.Data.GetCommands(common.DevfileOptions{
		CommandOptions: common.CommandOptions{
			CommandType: v1alpha2.CompositeCommandType,
		},
	})
	if err != nil {
		return err
	}

	for _, command := range commands {
		for _, subcommand := range command.Composite.Commands {
			if subcommand == name {
				return fmt.Errorf("command %q is used by composite command %q", name, command.Id)
			}
		}
	}

	return nil
}
