package exports

import (
	"syscall/js"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/feloy/devfile-builder/wasm/pkg/global"
	"github.com/feloy/devfile-builder/wasm/pkg/utils"
)

func AddContainerWrapper(this js.Value, args []js.Value) interface{} {
	command := getStringArray(args[2])
	arg := getStringArray(args[3])
	return result(
		addContainer(args[0].String(), args[1].String(), command, arg, args[4].String(), args[5].String(), args[6].String(), args[7].String()),
	)
}

func addContainer(name string, image string, command []string, args []string, memRequest string, memLimit string, cpuRequest string, cpuLimit string) (map[string]interface{}, error) {
	container := v1alpha2.Component{
		Name: name,
		ComponentUnion: v1alpha2.ComponentUnion{
			Container: &v1alpha2.ContainerComponent{
				Container: v1alpha2.Container{
					Image:         image,
					Command:       command,
					Args:          args,
					MemoryRequest: memRequest,
					MemoryLimit:   memLimit,
					CpuRequest:    cpuRequest,
					CpuLimit:      cpuLimit,
				},
			},
		},
	}
	err := global.Devfile.Data.AddComponents([]v1alpha2.Component{container})
	if err != nil {
		return nil, err
	}
	return utils.GetContent()
}
