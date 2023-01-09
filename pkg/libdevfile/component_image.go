package libdevfile

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
)

// imageComponent implements the component interface
type imageComponent struct {
	component  v1alpha2.Component
	devfileObj parser.DevfileObj
}

var _ component = (*imageComponent)(nil)

func newImageComponent(devfileObj parser.DevfileObj, component v1alpha2.Component) *imageComponent {
	return &imageComponent{
		component:  component,
		devfileObj: devfileObj,
	}
}

func (e *imageComponent) CheckValidity() error {
	return nil
}

func (e *imageComponent) Apply(handler Handler) error {
	return handler.ApplyImage(e.component)
}
