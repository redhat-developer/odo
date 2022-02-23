package libdevfile

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

type component interface {
	CheckValidity() error
	Apply(handler Handler) error
}

// newComponent creates a concrete component, based on its type
func newComponent(devfileObj parser.DevfileObj, devfileCmp v1alpha2.Component) (component, error) {
	var cmp component

	componentType, err := common.GetComponentType(devfileCmp)
	if err != nil {
		return nil, err
	}
	switch componentType {

	case v1alpha2.ContainerComponentType:
		cmp = newContainerComponent(devfileObj, devfileCmp)

	case v1alpha2.KubernetesComponentType:
		cmp = newKubernetesComponent(devfileObj, devfileCmp)

	case v1alpha2.OpenshiftComponentType:
		cmp = newOpenshiftComponent(devfileObj, devfileCmp)

	case v1alpha2.VolumeComponentType:
		cmp = newVolumeComponent(devfileObj, devfileCmp)

	case v1alpha2.ImageComponentType:
		cmp = newImageComponent(devfileObj, devfileCmp)
	}

	if err := cmp.CheckValidity(); err != nil {
		return nil, err
	}
	return cmp, nil
}
