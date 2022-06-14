package api

import (
	"github.com/redhat-developer/odo/pkg/kclient"
	bindingApi "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	specApi "github.com/redhat-developer/service-binding-operator/apis/spec/v1alpha3"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ServiceBinding describes a service binding, from group binding.operators.coreos.com/v1alpha1 or servicebinding.io/v1alpha3
type ServiceBinding struct {
	Name   string                `json:"name"`
	Spec   ServiceBindingSpec    `json:"spec"`
	Status *ServiceBindingStatus `json:"status,omitempty"`
}

type ServiceBindingSpec struct {
	Application            corev1.ObjectReference   `json:"application"`
	Services               []corev1.ObjectReference `json:"services"`
	DetectBindingResources bool                     `json:"detectBindingResources"`
	BindAsFiles            bool                     `json:"bindAsFiles"`
}

type ServiceBindingStatus struct {
	BindingFiles   []string        `json:"bindingFiles,omitempty"`
	BindingEnvVars []string        `json:"bindingEnvVars,omitempty"`
	RunningIn      RunningModeList `json:"runningIn,omitempty"`
}

// ServiceBindingFromBinding returns a common api.ServiceBinding structure
// from a ServiceBinding.binding.operators.coreos.com/v1alpha1
func ServiceBindingFromBinding(client kclient.ClientInterface, binding bindingApi.ServiceBinding) (ServiceBinding, error) {

	var dstSvcs []corev1.ObjectReference
	for _, srcSvc := range binding.Spec.Services {
		dstSvc := corev1.ObjectReference{
			Name: srcSvc.Name,
		}
		dstSvc.APIVersion, dstSvc.Kind = schema.GroupVersion{
			Group:   srcSvc.Group,
			Version: srcSvc.Version,
		}.WithKind(srcSvc.Kind).ToAPIVersionAndKind()
		dstSvcs = append(dstSvcs, dstSvc)
	}

	application := binding.Spec.Application
	refToApplication := corev1.ObjectReference{
		Name: application.Name,
	}

	if application.Kind == "" {
		gvk, err := client.GetGVKFromGVR(schema.GroupVersionResource{
			Group:    application.Group,
			Version:  application.Version,
			Resource: application.Resource,
		})
		if err != nil {
			return ServiceBinding{}, err
		}
		application.Kind = gvk.Kind
	}
	refToApplication.APIVersion, refToApplication.Kind = schema.GroupVersion{
		Group:   application.Group,
		Version: application.Version,
	}.WithKind(application.Kind).ToAPIVersionAndKind()

	return ServiceBinding{
		Name: binding.Name,
		Spec: ServiceBindingSpec{
			Application:            refToApplication,
			Services:               dstSvcs,
			DetectBindingResources: binding.Spec.DetectBindingResources,
			BindAsFiles:            binding.Spec.BindAsFiles,
		},
	}, nil
}

// ServiceBindingFromSpec returns a common api.ServiceBinding structure
// from a ServiceBinding.servicebinding.io/v1alpha3
func ServiceBindingFromSpec(spec specApi.ServiceBinding) ServiceBinding {

	service := spec.Spec.Service
	refToService := corev1.ObjectReference{
		APIVersion: service.APIVersion,
		Kind:       service.Kind,
		Name:       service.Name,
	}

	application := spec.Spec.Workload
	refToApplication := corev1.ObjectReference{
		APIVersion: application.APIVersion,
		Kind:       application.Kind,
		Name:       application.Name,
	}

	return ServiceBinding{
		Name: spec.Name,
		Spec: ServiceBindingSpec{
			Application:            refToApplication,
			Services:               []corev1.ObjectReference{refToService},
			DetectBindingResources: false,
			BindAsFiles:            true,
		},
	}
}
