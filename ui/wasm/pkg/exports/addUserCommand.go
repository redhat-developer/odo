package exports

import (
	"syscall/js"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/feloy/devfile-builder/wasm/pkg/global"
	"github.com/feloy/devfile-builder/wasm/pkg/utils"
)

func AddUserCommandWrapper(this js.Value, args []js.Value) interface{} {
	return result(
		addUserCommand(args[0].String(), args[1].String(), args[2].String()),
	)
}

func addUserCommand(component string, name string, commandLine string) (map[string]interface{}, error) {
	command := v1alpha2.Command{
		Id: name,
		CommandUnion: v1alpha2.CommandUnion{
			Exec: &v1alpha2.ExecCommand{
				CommandLine: commandLine,
				Component:   component,
			},
		},
	}
	err := global.Devfile.Data.AddCommands([]v1alpha2.Command{command})
	if err != nil {
		return nil, err
	}
	return utils.GetContent()
}
