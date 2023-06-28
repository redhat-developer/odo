package exports

import (
	"fmt"
	"syscall/js"

	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"

	"github.com/feloy/devfile-builder/wasm/pkg/exports/sub"
	"github.com/feloy/devfile-builder/wasm/pkg/global"
	"github.com/feloy/devfile-builder/wasm/pkg/utils"
)

func MoveCommandWrapper(this js.Value, args []js.Value) interface{} {
	return result(
		moveCommand(args[0].String(), args[1].String(), args[2].Int(), args[3].Int()),
	)
}

func moveCommand(previousGroup, newGroup string, previousIndex, newIndex int) (map[string]interface{}, error) {
	commands, err := global.Devfile.Data.GetCommands(common.DevfileOptions{})
	if err != nil {
		return nil, err
	}

	commandsByGroup, err := sub.MoveCommand(commands, previousGroup, newGroup, previousIndex, newIndex)
	if err != nil {
		return nil, err
	}

	// Deleting from the end as deleting from the beginning seems buggy
	for i := len(commands) - 1; i >= 0; i-- {
		global.Devfile.Data.DeleteCommand(commands[i].Id)
	}

	for _, group := range []string{"build", "run", "test", "debug", "deploy", ""} {
		err := global.Devfile.Data.AddCommands(commandsByGroup[group])
		if err != nil {
			fmt.Printf("%s\n", err)
			return nil, err
		}
	}

	return utils.GetContent()

}
