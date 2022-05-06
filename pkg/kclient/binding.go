package kclient

import (
	"context"
	"errors"

	sboApi "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	ServiceBindingGroup    = "binding.operators.coreos.com"
	ServiceBindingVersion  = "v1alpha1"
	ServiceBindingKind     = "ServiceBinding"
	ServiceBindingResource = "servicebindings"
	BindableKindsResource  = "bindablekinds"
)

// IsServiceBindingSupported checks if resource of type service binding request present on the cluster
func (c *Client) IsServiceBindingSupported() (bool, error) {
	// Detection of SBO has been removed from issue https://github.com/redhat-developer/odo/issues/5084
	return c.IsResourceSupported(ServiceBindingGroup, ServiceBindingVersion, ServiceBindingResource)
}

// GetBindableKinds returns BindableKind of name "bindable-kinds".
// "bindable-kinds" is the default resource provided by SBO
func (c *Client) GetBindableKinds() (sboApi.BindableKinds, error) {
	if c.DynamicClient == nil {
		return sboApi.BindableKinds{}, nil
	}

	var (
		unstructuredBK *unstructured.Unstructured
		bindableKind   sboApi.BindableKinds
		err            error
	)

	gvr := schema.GroupVersionResource{Group: ServiceBindingGroup, Version: ServiceBindingVersion, Resource: BindableKindsResource}
	unstructuredBK, err = c.DynamicClient.Resource(gvr).Get(context.TODO(), "bindable-kinds", v1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return sboApi.BindableKinds{}, errors.New("Service Binding Operator is not installed, please install it before proceeding")
		}
		return sboApi.BindableKinds{}, err
	}

	err = c.ConvertUnstructuredToResource(unstructuredBK.UnstructuredContent(), &bindableKind)
	if err != nil {
		return sboApi.BindableKinds{}, err
	}
	return bindableKind, nil
}

// GetBindableKindStatusRestMapping retuns a list of *meta.RESTMapping of all the bindable kind operator CRD
func (c Client) GetBindableKindStatusRestMapping(bindableKindStatuses []sboApi.BindableKindsStatus) ([]*meta.RESTMapping, error) {
	var bindableObjectRESTMappings []*meta.RESTMapping
	for _, bks := range bindableKindStatuses {
		if isBindableKindStatusGKAlreadyAdded(bindableObjectRESTMappings, bks) {
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
		bindableObjectRESTMappings = append(bindableObjectRESTMappings, restMapping)
	}
	return bindableObjectRESTMappings, nil
}

// NewServiceBindingServiceObject returns the sboApi.Service object based on the RESTMapping
func (c *Client) NewServiceBindingServiceObject(serviceRESTMapping *meta.RESTMapping, bindingName string, serviceName string) sboApi.Service {
	return sboApi.Service{
		Id: &bindingName, // Id field is helpful if user wants to inject mappings (custom binding data)
		NamespacedRef: sboApi.NamespacedRef{
			Ref: sboApi.Ref{
				Group:    serviceRESTMapping.GroupVersionKind.Group,
				Version:  serviceRESTMapping.GroupVersionKind.Version,
				Kind:     serviceRESTMapping.GroupVersionKind.Kind,
				Name:     serviceName,
				Resource: serviceRESTMapping.Resource.Resource,
			},
		},
	}
}

// NewServiceBindingObject returns the sboApi.ServiceBinding object
func (c *Client) NewServiceBindingObject(bindingName string, bindAsFiles bool, deploymentName string, deploymentGVR v1.GroupVersionResource, mappings []sboApi.Mapping, services []sboApi.Service) *sboApi.ServiceBinding {
	return &sboApi.ServiceBinding{
		TypeMeta: v1.TypeMeta{
			APIVersion: schema.GroupVersion{
				Group:   ServiceBindingGroup,
				Version: ServiceBindingVersion,
			}.String(),
			Kind: ServiceBindingKind,
		},
		ObjectMeta: v1.ObjectMeta{
			Name: bindingName,
		},
		Spec: sboApi.ServiceBindingSpec{
			DetectBindingResources: true,
			BindAsFiles:            bindAsFiles,
			Application: sboApi.Application{
				Ref: sboApi.Ref{
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

func isBindableKindStatusGKAlreadyAdded(bindableObjects []*meta.RESTMapping, bks sboApi.BindableKindsStatus) (gkAlreadyAdded bool) {
	// check every GroupKind only once
	for _, bo := range bindableObjects {
		if bo.GroupVersionKind.Group == bks.Group && bo.GroupVersionKind.Kind == bks.Kind {
			gkAlreadyAdded = true
			break
		}
	}
	return
}
