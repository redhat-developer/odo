package exports

import (
	"syscall/js"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/feloy/devfile-builder/wasm/pkg/global"
	"github.com/feloy/devfile-builder/wasm/pkg/utils"
)

func AddExecCommandWrapper(this js.Value, args []js.Value) interface{} {
	return result(
		addExecCommand(args[0].String(), args[1].String(), args[2].String(), args[3].String(), args[4].Bool()),
	)
}

func addExecCommand(name string, component string, commandLine string, workingDir string, hotReloadCapable bool) (map[string]interface{}, error) {
	command := v1alpha2.Command{
		Id: name,
		CommandUnion: v1alpha2.CommandUnion{
			Exec: &v1alpha2.ExecCommand{
				Component:        component,
				CommandLine:      commandLine,
				WorkingDir:       workingDir,
				HotReloadCapable: &hotReloadCapable,
			},
		},
	}
	err := global.Devfile.Data.AddCommands([]v1alpha2.Command{command})
	if err != nil {
		return nil, err
	}
	return utils.GetContent()
}
