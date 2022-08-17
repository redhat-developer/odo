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

type ImageComponentParams struct {
	Name       string
	Attributes *attributes.Attributes

	Image v1alpha2.Image
}

func GetImageComponent(params ImageComponentParams) v1alpha2.Component {
	cmp := v1alpha2.Component{
		Name: params.Name,
		ComponentUnion: v1alpha2.ComponentUnion{
			Image: &v1alpha2.ImageComponent{
				Image: params.Image,
			},
		},
	}
	if params.Attributes != nil {
		cmp.Attributes = *params.Attributes
	}
	return cmp
}

type KubernetesComponentParams struct {
	Name       string
	Attributes *attributes.Attributes

	Kubernetes *v1alpha2.KubernetesComponent
}

func GetKubernetesComponent(params KubernetesComponentParams) v1alpha2.Component {
	cmp := v1alpha2.Component{
		Name: params.Name,
		ComponentUnion: v1alpha2.ComponentUnion{
			Kubernetes: params.Kubernetes,
		},
	}
	if params.Attributes != nil {
		cmp.Attributes = *params.Attributes
	}
	return cmp
}

type OpenshiftComponentParams struct {
	Name       string
	Attributes *attributes.Attributes

	Openshift *v1alpha2.OpenshiftComponent
}

func GetOpenshiftComponent(params OpenshiftComponentParams) v1alpha2.Component {
	cmp := v1alpha2.Component{
		Name: params.Name,
		ComponentUnion: v1alpha2.ComponentUnion{
			Openshift: params.Openshift,
		},
	}
	if params.Attributes != nil {
		cmp.Attributes = *params.Attributes
	}
	return cmp
}
