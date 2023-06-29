package exports

import (
	"fmt"
	"syscall/js"

	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"

	"github.com/feloy/devfile-builder/wasm/pkg/global"
	"github.com/feloy/devfile-builder/wasm/pkg/utils"
)

func UnsetDefaultCommandWrapper(this js.Value, args []js.Value) interface{} {
	return result(
		unsetDefaultCommand(args[0].String()),
	)
}

func unsetDefaultCommand(commandName string) (map[string]interface{}, error) {
	commands, err := global.Devfile.Data.GetCommands(common.DevfileOptions{})
	if err != nil {
		return nil, err
	}

	for _, command := range commands {
		if command.Id == commandName {
			utils.SetDefault(&command, false)
			err = global.Devfile.Data.UpdateCommand(command)
			if err != nil {
				fmt.Printf("%s\n", err)
				return nil, err
			}
			break
		}
	}
	return utils.GetContent()
}
