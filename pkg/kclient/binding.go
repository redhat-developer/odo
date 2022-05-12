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
	ServiceBindingKind    = "ServiceBinding"
	BindableKindsResource = "bindablekinds"
)

// IsServiceBindingSupported checks if resource of type service binding request present on the cluster
func (c *Client) IsServiceBindingSupported() (bool, error) {
	gvr := sboApi.GroupVersionResource
	return c.IsResourceSupported(gvr.Group, gvr.Version, gvr.Resource)
}

// GetBindableKinds returns BindableKinds of name "bindable-kinds".
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

	unstructuredBK, err = c.DynamicClient.Resource(sboApi.GroupVersion.WithResource(BindableKindsResource)).Get(context.TODO(), "bindable-kinds", v1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			//revive:disable:error-strings This is a top-level error message displayed as is to the end user
			return bindableKind, errors.New("Service Binding Operator is not installed or it is not completely installed. Please ensure that it is installed successfully before proceeding.")
			//revive:enable:error-strings
		}
		return bindableKind, err
	}

	err = c.ConvertUnstructuredToResource(unstructuredBK.UnstructuredContent(), &bindableKind)
	if err != nil {
		return bindableKind, err
	}
	return bindableKind, nil
}

// GetBindableKindStatusRestMapping retuns a list of *meta.RESTMapping of all the bindable kind operator CRD
func (c Client) GetBindableKindStatusRestMapping(bindableKindStatuses []sboApi.BindableKindsStatus) ([]*meta.RESTMapping, error) {
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

// NewServiceBindingServiceObject returns the sboApi.Service object based on the RESTMapping
func (c *Client) NewServiceBindingServiceObject(unstructuredService unstructured.Unstructured, bindingName string) (sboApi.Service, error) {
	serviceRESTMapping, err := c.GetRestMappingFromUnstructured(unstructuredService)
	if err != nil {
		return sboApi.Service{}, err
	}

	return sboApi.Service{
		Id: &bindingName, // Id field is helpful if user wants to inject mappings (custom binding data)
		NamespacedRef: sboApi.NamespacedRef{
			Ref: sboApi.Ref{
				Group:    serviceRESTMapping.GroupVersionKind.Group,
				Version:  serviceRESTMapping.GroupVersionKind.Version,
				Kind:     serviceRESTMapping.GroupVersionKind.Kind,
				Name:     unstructuredService.GetName(),
				Resource: serviceRESTMapping.Resource.Resource,
			},
		},
	}, nil
}

// NewServiceBindingObject returns the sboApi.ServiceBinding object
func NewServiceBindingObject(bindingName string, bindAsFiles bool, deploymentName string, deploymentGVR v1.GroupVersionResource, mappings []sboApi.Mapping, services []sboApi.Service) *sboApi.ServiceBinding {
	return &sboApi.ServiceBinding{
		TypeMeta: v1.TypeMeta{
			APIVersion: sboApi.GroupVersion.String(),
			Kind:       ServiceBindingKind,
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

func mappingContainsBKS(bindableObjects []*meta.RESTMapping, bks sboApi.BindableKindsStatus) bool {
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
