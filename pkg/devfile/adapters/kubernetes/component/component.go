package component

import (
	"fmt"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
)

// componentToApply represents a devfile component that can be applied
type componentToApply interface {
	Apply(devfileObj parser.DevfileObj, devfilePath string) error
}

// createComponent returns an instance of a devfile component specific to its type (image, kubernetes, etc)
func createComponent(adapter Adapter, component devfilev1.Component) (componentToApply, error) {
	if component.Image != nil {
		return newComponentImage(component), nil
	} else if component.Kubernetes != nil {
		return newComponentKubernetes(adapter.Client, component, adapter.ComponentName, adapter.AppName), nil
	}
	return nil, fmt.Errorf("component type not supported for component %q", component.Name)
}
