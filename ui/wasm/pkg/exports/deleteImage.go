package exports

import (
	"fmt"
	"syscall/js"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"

	"github.com/feloy/devfile-builder/wasm/pkg/global"
	"github.com/feloy/devfile-builder/wasm/pkg/utils"
)

func DeleteImageWrapper(this js.Value, args []js.Value) interface{} {
	return result(
		deleteImage(args[0].String()),
	)
}

func deleteImage(name string) (map[string]interface{}, error) {

	err := checkImageUsed(name)
	if err != nil {
		return nil, fmt.Errorf("error deleting image %q: %w", name, err)
	}
	err = global.Devfile.Data.DeleteComponent(name)
	if err != nil {
		return nil, err
	}
	return utils.GetContent()
}

func checkImageUsed(name string) error {
	commands, err := global.Devfile.Data.GetCommands(common.DevfileOptions{
		CommandOptions: common.CommandOptions{
			CommandType: v1alpha2.ApplyCommandType,
		},
	})
	if err != nil {
		return err
	}

	for _, command := range commands {
		if command.Apply.Component == name {
			return fmt.Errorf("image %q is used by Image Command %q", name, command.Id)
		}
	}

	return nil
}
