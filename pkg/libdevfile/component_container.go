package libdevfile

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
)

// containerComponent implements the component interface
type containerComponent struct {
	component  v1alpha2.Component
	devfileObj parser.DevfileObj
}

var _ component = (*containerComponent)(nil)

func newContainerComponent(devfileObj parser.DevfileObj, component v1alpha2.Component) *containerComponent {
	return &containerComponent{
		component:  component,
		devfileObj: devfileObj,
	}
}

func (e *containerComponent) CheckValidity() error {
	return nil
}

func (e *containerComponent) Apply(handler Handler) error {
	return nil
}
