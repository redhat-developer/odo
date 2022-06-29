package libdevfile

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
)

// kubernetesComponent implements the component interface
type kubernetesComponent struct {
	component  v1alpha2.Component
	devfileObj parser.DevfileObj
}

var _ component = (*kubernetesComponent)(nil)

func newKubernetesComponent(devfileObj parser.DevfileObj, component v1alpha2.Component) *kubernetesComponent {
	return &kubernetesComponent{
		component:  component,
		devfileObj: devfileObj,
	}
}

func (e *kubernetesComponent) CheckValidity() error {
	return nil
}

func (e *kubernetesComponent) Apply(handler Handler) error {
	return handler.ApplyKubernetes(e.component)
}
