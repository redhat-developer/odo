package exports

import (
	"fmt"
	"syscall/js"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"

	"github.com/feloy/devfile-builder/wasm/pkg/global"
	"github.com/feloy/devfile-builder/wasm/pkg/utils"
)

func DeleteContainerWrapper(this js.Value, args []js.Value) interface{} {
	return result(
		deleteContainer(args[0].String()),
	)
}

func deleteContainer(name string) (map[string]interface{}, error) {

	err := checkContainerUsed(name)
	if err != nil {
		return nil, fmt.Errorf("error deleting container %q: %w", name, err)
	}
	err = global.Devfile.Data.DeleteComponent(name)
	if err != nil {
		return nil, err
	}
	return utils.GetContent()
}

func checkContainerUsed(name string) error {
	commands, err := global.Devfile.Data.GetCommands(common.DevfileOptions{
		CommandOptions: common.CommandOptions{
			CommandType: v1alpha2.ExecCommandType,
		},
	})
	if err != nil {
		return err
	}

	for _, command := range commands {
		if command.Exec.Component == name {
			return fmt.Errorf("container %q is used by exec command %q", name, command.Id)
		}
	}

	return nil
}
