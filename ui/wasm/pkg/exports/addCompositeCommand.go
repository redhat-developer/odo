package exports

import (
	"syscall/js"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/feloy/devfile-builder/wasm/pkg/global"
	"github.com/feloy/devfile-builder/wasm/pkg/utils"
)

func AddCompositeCommandWrapper(this js.Value, args []js.Value) interface{} {
	return result(
		addCompositeCommand(args[0].String(), args[1].Bool(), getStringArray(args[2])),
	)
}

func addCompositeCommand(name string, parallel bool, commands []string) (map[string]interface{}, error) {
	command := v1alpha2.Command{
		Id: name,
		CommandUnion: v1alpha2.CommandUnion{
			Composite: &v1alpha2.CompositeCommand{
				Parallel: &parallel,
				Commands: commands,
			},
		},
	}
	err := global.Devfile.Data.AddCommands([]v1alpha2.Command{command})
	if err != nil {
		return nil, err
	}
	return utils.GetContent()
}
