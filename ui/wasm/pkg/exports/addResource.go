package exports

import (
	"errors"
	"syscall/js"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/feloy/devfile-builder/wasm/pkg/global"
	"github.com/feloy/devfile-builder/wasm/pkg/utils"
)

func AddResourceWrapper(this js.Value, args []js.Value) interface{} {
	return result(
		addResource(args[0].String(), args[1].String(), args[2].String()),
	)
}

func addResource(name string, inlined string, uri string) (map[string]interface{}, error) {
	if inlined != "" && uri != "" {
		return nil, errors.New("both inlined and uri cannot be set at the same time")
	}
	container := v1alpha2.Component{
		Name: name,
		ComponentUnion: v1alpha2.ComponentUnion{
			Kubernetes: &v1alpha2.KubernetesComponent{
				K8sLikeComponent: v1alpha2.K8sLikeComponent{
					K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
						Inlined: inlined,
						Uri:     uri,
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
