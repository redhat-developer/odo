package api

import (
	bindingApi "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	specApi "github.com/redhat-developer/service-binding-operator/apis/spec/v1alpha3"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ServiceBinding describes a service binding, from group binding.operators.coreos.com/v1alpha1 or servicebinding.io/v1alpha3
type ServiceBinding struct {
	Name   string                `json:"name"`
	Spec   ServiceBindingSpec    `json:"spec"`
	Status *ServiceBindingStatus `json:"status,omitempty"`
}

type ServiceBindingSpec struct {
	Services               []specApi.ServiceBindingServiceReference `json:"services"`
	DetectBindingResources bool                                     `json:"detectBindingResources"`
	BindAsFiles            bool                                     `json:"bindAsFiles"`
}

type ServiceBindingStatus struct {
	BindingFiles   []string `json:"bindingsFiles,omitempty"`
	BindingEnvVars []string `json:"bindingEnvVars,omitempty"`
}

// ServiceBindingFromBinding returns a common api.ServiceBinding structure
// from a ServiceBinding.binding.operators.coreos.com/v1alpha1
func ServiceBindingFromBinding(binding bindingApi.ServiceBinding) ServiceBinding {

	var dstSvcs []specApi.ServiceBindingServiceReference
	for _, srcSvc := range binding.Spec.Services {
		dstSvc := specApi.ServiceBindingServiceReference{
			Name: srcSvc.Name,
		}
		dstSvc.APIVersion, dstSvc.Kind = schema.GroupVersion{
			Group:   srcSvc.Group,
			Version: srcSvc.Version,
		}.WithKind(srcSvc.Kind).ToAPIVersionAndKind()
		dstSvcs = append(dstSvcs, dstSvc)
	}
	return ServiceBinding{
		Name: binding.Name,
		Spec: ServiceBindingSpec{
			Services:               dstSvcs,
			DetectBindingResources: binding.Spec.DetectBindingResources,
			BindAsFiles:            binding.Spec.BindAsFiles,
		},
	}
}

// ServiceBindingFromSpec returns a common api.ServiceBinding structure
// from a ServiceBinding.servicebinding.io/v1alpha3
func ServiceBindingFromSpec(spec specApi.ServiceBinding) ServiceBinding {
	return ServiceBinding{
		Name: spec.Name,
		Spec: ServiceBindingSpec{
			Services:               []specApi.ServiceBindingServiceReference{spec.Spec.Service},
			DetectBindingResources: false,
			BindAsFiles:            true,
		},
	}
}
