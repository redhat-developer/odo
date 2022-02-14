package generator

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
)

type ContainerComponentParams struct {
	Name       string
	Attributes *attributes.Attributes

	Container v1alpha2.Container
	Endpoints []v1alpha2.Endpoint
}

func GetContainerComponent(params ContainerComponentParams) v1alpha2.Component {
	cmp := v1alpha2.Component{
		Name: params.Name,
		ComponentUnion: v1alpha2.ComponentUnion{
			Container: &v1alpha2.ContainerComponent{
				Container: params.Container,
				Endpoints: params.Endpoints,
			},
		},
	}
	if params.Attributes != nil {
		cmp.Attributes = *params.Attributes
	}
	return cmp
}
