package component

import (
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/image"
)

// componentImage represents a devfile component of type Image
type componentImage struct {
	component devfilev1.Component
}

func newComponentImage(component devfilev1.Component) componentImage {
	return componentImage{component: component}
}

// Apply a component of type Image by building and pushing the image
func (o componentImage) Apply(devfileObj parser.DevfileObj, devfilePath string) error {
	return image.BuildPushSpecificImage(devfileObj, devfilePath, o.component, true)
}
