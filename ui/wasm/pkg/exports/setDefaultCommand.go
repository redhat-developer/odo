package exports

import (
	"fmt"
	"syscall/js"

	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"

	"github.com/feloy/devfile-builder/wasm/pkg/global"
	"github.com/feloy/devfile-builder/wasm/pkg/utils"
)

func SetDefaultCommandWrapper(this js.Value, args []js.Value) interface{} {
	return result(
		setDefaultCommand(args[0].String(), args[1].String()),
	)
}

func setDefaultCommand(commandName string, group string) (map[string]interface{}, error) {
	fmt.Printf("change default for group %q\n", group)
	commands, err := global.Devfile.Data.GetCommands(common.DevfileOptions{})
	if err != nil {
		return nil, err
	}

	for _, command := range commands {
		if utils.GetGroup(command) == group {
			isDefault := command.Id == commandName
			fmt.Printf("setting default = %v for command %q\n", isDefault, command.Id)
			utils.SetDefault(&command, isDefault)
			err = global.Devfile.Data.UpdateCommand(command)
			if err != nil {
				fmt.Printf("%s\n", err)
				return nil, err
			}
		}
	}
	return utils.GetContent()
}
