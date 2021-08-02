package context

import (
	"context"
	"fmt"
	"github.com/redhat-developer/service-binding-operator/apis/spec/v1alpha2"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
)

var _ pipeline.Context = &specImpl{}

var SpecProvider = func(client dynamic.Interface, typeLookup K8STypeLookup) pipeline.ContextProvider {
	return &provider{
		client:     client,
		typeLookup: typeLookup,
		get: func(binding interface{}) (pipeline.Context, error) {
			switch sb := binding.(type) {
			case *v1alpha2.ServiceBinding:
				if sb.Generation != 0 {
					sb.Status.ObservedGeneration = sb.Generation
				}
				ctx := &specImpl{
					impl: impl{
						conditions:  make(map[string]*metav1.Condition),
						client:      client,
						typeLookup:  typeLookup,
						bindingMeta: &sb.ObjectMeta,
						statusSecretName: func() string {
							if sb.Status.Binding == nil {
								return ""
							}
							return sb.Status.Binding.Name
						},
						setStatusSecretName: func(name string) {
							sb.Status.Binding = &v1alpha2.ServiceBindingSecretReference{Name: name}
						},
						unstructuredBinding: func() (*unstructured.Unstructured, error) {
							return converter.ToUnstructured(sb)
						},
						statusConditions: func() *[]metav1.Condition {
							return &sb.Status.Conditions
						},
						ownerReference: func() metav1.OwnerReference {
							return sb.AsOwnerReference()
						},
						groupVersionResource: func() schema.GroupVersionResource {
							return v1alpha2.GroupVersionResource
						},
					},
					serviceBinding: sb,
				}
				if sb.Spec.Type != "" {
					ctx.AddBindingItem(&pipeline.BindingItem{Name: "type", Value: sb.Spec.Type})
				}
				if sb.Spec.Provider != "" {
					ctx.AddBindingItem(&pipeline.BindingItem{Name: "provider", Value: sb.Spec.Provider})
				}
				return ctx, nil
			}
			return nil, fmt.Errorf("cannot create context for passed instance %v", binding)
		},
	}
}

type specImpl struct {
	impl
	serviceBinding *v1alpha2.ServiceBinding
}

func (i *specImpl) BindingName() string {
	if i.serviceBinding.Spec.Name != "" {
		return i.serviceBinding.Spec.Name
	}
	return i.bindingMeta.Name
}

func (i *specImpl) EnvBindings() []*pipeline.EnvBinding {
	if len(i.serviceBinding.Spec.Env) == 0 {
		return make([]*pipeline.EnvBinding, 0)
	}
	result := make([]*pipeline.EnvBinding, 0, len(i.serviceBinding.Spec.Env))
	for _, e := range i.serviceBinding.Spec.Env {
		result = append(result, &pipeline.EnvBinding{Var: e.Name, Name: e.Key})
	}
	return result
}

func (i *specImpl) Services() ([]pipeline.Service, error) {
	if i.services == nil {
		serviceRef := i.serviceBinding.Spec.Service

		gvr, err := i.typeLookup.ResourceForReferable(&serviceRef)
		if err != nil {
			return nil, err
		}
		u, err := i.client.Resource(*gvr).Namespace(i.serviceBinding.Namespace).Get(context.Background(), serviceRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		i.services = append(i.services, &service{client: i.client, resource: u, groupVersionResource: gvr, namespace: i.serviceBinding.Namespace})
	}
	services := make([]pipeline.Service, len(i.services))
	for idx := 0; idx < len(i.services); idx++ {
		services[idx] = i.services[idx]
	}
	return services, nil

}

func (i *specImpl) Applications() ([]pipeline.Application, error) {
	if i.applications == nil {
		ref := i.serviceBinding.Spec.Application
		gvr, err := i.typeLookup.ResourceForReferable(&ref)
		if err != nil {
			return nil, err
		}
		if i.serviceBinding.Spec.Application.Name != "" {
			u, err := i.client.Resource(*gvr).Namespace(i.serviceBinding.Namespace).Get(context.Background(), ref.Name, metav1.GetOptions{})
			if err != nil {
				return nil, emptyApplicationsErr{err}
			}
			i.applications = append(i.applications, &application{gvr: gvr, persistedResource: u, bindableContainerNames: sets.NewString(i.serviceBinding.Spec.Application.Containers...)})
		}
		if i.serviceBinding.Spec.Application.Selector.MatchLabels != nil {
			matchLabels := i.serviceBinding.Spec.Application.Selector.MatchLabels
			opts := metav1.ListOptions{
				LabelSelector: labels.Set(matchLabels).String(),
			}

			objList, err := i.client.Resource(*gvr).Namespace(i.serviceBinding.Namespace).List(context.Background(), opts)
			if err != nil {
				return nil, err
			}

			if len(objList.Items) == 0 {
				return nil, emptyApplicationsErr{}
			}

			for index := range objList.Items {
				i.applications = append(i.applications, &application{gvr: gvr, persistedResource: &(objList.Items[index]), bindableContainerNames: sets.NewString(i.serviceBinding.Spec.Application.Containers...)})
			}
		}
	}

	result := make([]pipeline.Application, len(i.applications))
	for l, a := range i.applications {
		result[l] = a
	}
	return result, nil

}

func (s *specImpl) BindAsFiles() bool {
	return true
}

func (s *specImpl) MountPath() string {
	return ""
}

func (s *specImpl) NamingTemplate() string {
	return ""
}

func (s *specImpl) Mappings() map[string]string {
	return make(map[string]string)
}
