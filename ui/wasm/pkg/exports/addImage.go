package exports

import (
	"syscall/js"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/feloy/devfile-builder/wasm/pkg/global"
	"github.com/feloy/devfile-builder/wasm/pkg/utils"
)

func AddImageWrapper(this js.Value, args []js.Value) interface{} {
	arg := getStringArray(args[2])
	return result(
		addImage(args[0].String(), args[1].String(), arg, args[3].String(), args[4].Bool(), args[5].String()),
	)
}

func addImage(name string, imageName string, args []string, buildContext string, rootRequired bool, uri string) (map[string]interface{}, error) {
	container := v1alpha2.Component{
		Name: name,
		ComponentUnion: v1alpha2.ComponentUnion{
			Image: &v1alpha2.ImageComponent{
				Image: v1alpha2.Image{
					ImageName: imageName,
					ImageUnion: v1alpha2.ImageUnion{
						Dockerfile: &v1alpha2.DockerfileImage{
							Dockerfile: v1alpha2.Dockerfile{
								Args:         args,
								BuildContext: buildContext,
								RootRequired: &rootRequired,
							},
							DockerfileSrc: v1alpha2.DockerfileSrc{
								Uri: uri,
							},
						},
					},
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
