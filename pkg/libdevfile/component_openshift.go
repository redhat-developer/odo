package libdevfile

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
)

// openshiftComponent implements the component interface
type openshiftComponent struct {
	component  v1alpha2.Component
	devfileObj parser.DevfileObj
}

var _ component = (*openshiftComponent)(nil)

func newOpenshiftComponent(devfileObj parser.DevfileObj, component v1alpha2.Component) *openshiftComponent {
	return &openshiftComponent{
		component:  component,
		devfileObj: devfileObj,
	}
}

func (e *openshiftComponent) CheckValidity() error {
	return nil
}

func (e *openshiftComponent) Apply(handler Handler) error {
	return nil
}
