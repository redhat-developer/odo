package libdevfile

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
)

// volumeComponent implements the component interface
type volumeComponent struct {
	component  v1alpha2.Component
	devfileObj parser.DevfileObj
}

var _ component = (*volumeComponent)(nil)

func newVolumeComponent(devfileObj parser.DevfileObj, component v1alpha2.Component) *volumeComponent {
	return &volumeComponent{
		component:  component,
		devfileObj: devfileObj,
	}
}

func (e *volumeComponent) CheckValidity() error {
	return nil
}

func (e *volumeComponent) Apply(handler Handler) error {
	return nil
}
