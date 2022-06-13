package kclient

import (
	"context"
	"errors"

	bindingApi "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	specApi "github.com/redhat-developer/service-binding-operator/apis/spec/v1alpha3"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	ServiceBindingKind    = "ServiceBinding"
	BindableKindsResource = "bindablekinds"
)

// IsServiceBindingSupported checks if resource of type service binding request present on the cluster
func (c *Client) IsServiceBindingSupported() (bool, error) {
	gvr := bindingApi.GroupVersionResource
	return c.IsResourceSupported(gvr.Group, gvr.Version, gvr.Resource)
}

// GetBindableKinds returns BindableKinds of name "bindable-kinds".
// "bindable-kinds" is the default resource provided by SBO
func (c *Client) GetBindableKinds() (bindingApi.BindableKinds, error) {
	if c.DynamicClient == nil {
		return bindingApi.BindableKinds{}, nil
	}

	var (
		unstructuredBK *unstructured.Unstructured
		bindableKind   bindingApi.BindableKinds
		err            error
	)

	unstructuredBK, err = c.DynamicClient.Resource(bindingApi.GroupVersion.WithResource(BindableKindsResource)).Get(context.TODO(), "bindable-kinds", v1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			//revive:disable:error-strings This is a top-level error message displayed as is to the end user
			return bindableKind, errors.New("Service Binding Operator is not installed or it is not completely installed. Please ensure that it is installed successfully before proceeding.")
			//revive:enable:error-strings
		}
		return bindableKind, err
	}

	err = c.ConvertUnstructuredToResource(*unstructuredBK, &bindableKind)
	if err != nil {
		return bindableKind, err
	}
	return bindableKind, nil
}

// GetBindableKindStatusRestMapping returns a list of *meta.RESTMapping of all the bindable kind operator CRD
func (c Client) GetBindableKindStatusRestMapping(bindableKindStatuses []bindingApi.BindableKindsStatus) ([]*meta.RESTMapping, error) {
	var result []*meta.RESTMapping
	for _, bks := range bindableKindStatuses {
		if mappingContainsBKS(result, bks) {
			continue
		}
		restMapping, err := c.GetRestMappingFromGVK(schema.GroupVersionKind{
			Group:   bks.Group,
			Version: bks.Version,
			Kind:    bks.Kind,
		})
		if err != nil {
			return nil, err
		}
		result = append(result, restMapping)
	}
	return result, nil
}

// NewServiceBindingServiceObject returns the bindingApi.Service object based on the RESTMapping
func (c *Client) NewServiceBindingServiceObject(unstructuredService unstructured.Unstructured, bindingName string) (bindingApi.Service, error) {
	serviceRESTMapping, err := c.GetRestMappingFromUnstructured(unstructuredService)
	if err != nil {
		return bindingApi.Service{}, err
	}

	return bindingApi.Service{
		Id: &bindingName, // Id field is helpful if user wants to inject mappings (custom binding data)
		NamespacedRef: bindingApi.NamespacedRef{
			Ref: bindingApi.Ref{
				Group:    serviceRESTMapping.GroupVersionKind.Group,
				Version:  serviceRESTMapping.GroupVersionKind.Version,
				Kind:     serviceRESTMapping.GroupVersionKind.Kind,
				Name:     unstructuredService.GetName(),
				Resource: serviceRESTMapping.Resource.Resource,
			},
		},
	}, nil
}

// NewServiceBindingObject returns the bindingApi.ServiceBinding object
func NewServiceBindingObject(bindingName string, bindAsFiles bool, deploymentName string, deploymentGVR schema.GroupVersionResource, mappings []bindingApi.Mapping, services []bindingApi.Service) *bindingApi.ServiceBinding {
	return &bindingApi.ServiceBinding{
		TypeMeta: v1.TypeMeta{
			APIVersion: bindingApi.GroupVersion.String(),
			Kind:       ServiceBindingKind,
		},
		ObjectMeta: v1.ObjectMeta{
			Name: bindingName,
		},
		Spec: bindingApi.ServiceBindingSpec{
			DetectBindingResources: true,
			BindAsFiles:            bindAsFiles,
			Application: bindingApi.Application{
				Ref: bindingApi.Ref{
					Name:     deploymentName,
					Group:    deploymentGVR.Group,
					Version:  deploymentGVR.Version,
					Resource: deploymentGVR.Resource,
				},
			},
			Mappings: mappings,
			Services: services,
		},
	}
}

// GetBindingServiceBinding returns a ServiceBinding from group binding.operators.coreos.com/v1alpha1
func (c Client) GetBindingServiceBinding(name string) (bindingApi.ServiceBinding, error) {
	if c.DynamicClient == nil {
		return bindingApi.ServiceBinding{}, nil
	}

	u, err := c.GetDynamicResource(bindingApi.GroupVersionResource, name)
	if err != nil {
		return bindingApi.ServiceBinding{}, err
	}

	var result bindingApi.ServiceBinding
	err = c.ConvertUnstructuredToResource(*u, &result)
	if err != nil {
		return bindingApi.ServiceBinding{}, err
	}
	return result, nil
}

// GetSpecServiceBinding returns a ServiceBinding from group servicebinding.io/v1alpha3
func (c Client) GetSpecServiceBinding(name string) (specApi.ServiceBinding, error) {
	if c.DynamicClient == nil {
		return specApi.ServiceBinding{}, nil
	}

	u, err := c.GetDynamicResource(specApi.GroupVersionResource, name)
	if err != nil {
		return specApi.ServiceBinding{}, err
	}

	var result specApi.ServiceBinding
	err = c.ConvertUnstructuredToResource(*u, &result)
	if err != nil {
		return specApi.ServiceBinding{}, err
	}
	return result, nil
}

func mappingContainsBKS(bindableObjects []*meta.RESTMapping, bks bindingApi.BindableKindsStatus) bool {
	var gkAlreadyAdded bool
	// check every GroupKind only once
	for _, bo := range bindableObjects {
		if bo.GroupVersionKind.Group == bks.Group && bo.GroupVersionKind.Kind == bks.Kind {
			gkAlreadyAdded = true
			break
		}
	}
	return gkAlreadyAdded
}
