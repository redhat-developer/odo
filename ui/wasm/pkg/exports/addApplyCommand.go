package exports

import (
	"syscall/js"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/feloy/devfile-builder/wasm/pkg/global"
	"github.com/feloy/devfile-builder/wasm/pkg/utils"
)

func AddApplyCommandWrapper(this js.Value, args []js.Value) interface{} {
	return result(
		addApplyCommand(args[0].String(), args[1].String()),
	)
}

func addApplyCommand(name string, component string) (map[string]interface{}, error) {
	command := v1alpha2.Command{
		Id: name,
		CommandUnion: v1alpha2.CommandUnion{
			Apply: &v1alpha2.ApplyCommand{
				Component: component,
			},
		},
	}
	err := global.Devfile.Data.AddCommands([]v1alpha2.Command{command})
	if err != nil {
		return nil, err
	}
	return utils.GetContent()
}
