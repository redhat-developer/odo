package exports

import (
	"syscall/js"

	"github.com/feloy/devfile-builder/wasm/pkg/global"
	"github.com/feloy/devfile-builder/wasm/pkg/utils"
)

func UpdateEventsWrapper(this js.Value, args []js.Value) interface{} {
	commands := getStringArray(args[1])
	return result(
		updateEvents(args[0].String(), commands),
	)
}

func updateEvents(event string, commands []string) (map[string]interface{}, error) {
	switch event {
	case "postStart":
		global.Devfile.Data.UpdateEvents(commands, nil, nil, nil)
	case "postStop":
		global.Devfile.Data.UpdateEvents(nil, commands, nil, nil)
	case "preStart":
		global.Devfile.Data.UpdateEvents(nil, nil, commands, nil)
	case "preStop":
		global.Devfile.Data.UpdateEvents(nil, nil, nil, commands)
	}
	return utils.GetContent()
}
