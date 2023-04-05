package libdevfile

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
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

// GetK8sAndOcComponentsToPush returns the list of Kubernetes and OpenShift components to push,
// The list returned is governed by the DeployByDefault field in each component.
// All components with DeployByDefault set to true are included, along with those with no DeployByDefault set and not-referenced.
// It takes an additional allowApply boolean, which set to true, will append the components referenced from apply commands to the list.
func GetK8sAndOcComponentsToPush(devfileObj parser.DevfileObj, allowApply bool) ([]v1alpha2.Component, error) {
	var allK8sAndOcComponents []v1alpha2.Component

	k8sComponents, err := devfileObj.Data.GetComponents(parsercommon.DevfileOptions{
		ComponentOptions: parsercommon.ComponentOptions{ComponentType: v1alpha2.KubernetesComponentType},
	})
	if err != nil {
		return nil, err
	}
	allK8sAndOcComponents = append(allK8sAndOcComponents, k8sComponents...)

	ocComponents, err := devfileObj.Data.GetComponents(parsercommon.DevfileOptions{
		ComponentOptions: parsercommon.ComponentOptions{ComponentType: v1alpha2.OpenshiftComponentType},
	})
	if err != nil {
		return nil, err
	}
	allK8sAndOcComponents = append(allK8sAndOcComponents, ocComponents...)

	allApplyCommands, err := devfileObj.Data.GetCommands(parsercommon.DevfileOptions{
		CommandOptions: parsercommon.CommandOptions{CommandType: v1alpha2.ApplyCommandType},
	})
	if err != nil {
		return nil, err
	}

	m := make(map[string]v1alpha2.Component)
	for _, comp := range allK8sAndOcComponents {
		var k v1alpha2.K8sLikeComponent
		if comp.Kubernetes != nil {
			k = comp.Kubernetes.K8sLikeComponent
		} else {
			k = comp.Openshift.K8sLikeComponent
		}
		var add bool
		if allowApply && isComponentReferenced(allApplyCommands, comp.Name) {
			add = true
		} else if k.DeployByDefault == nil {
			// auto-created only if not referenced by any apply command
			if !isComponentReferenced(allApplyCommands, comp.Name) {
				add = true
			}
		} else if *k.DeployByDefault {
			add = true
		}
		if !add {
			continue
		}
		if _, present := m[comp.Name]; !present {
			m[comp.Name] = comp
		}
	}

	var result []v1alpha2.Component
	for _, comp := range m {
		result = append(result, comp)
	}
	return result, nil
}
