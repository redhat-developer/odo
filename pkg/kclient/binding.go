package kclient

import (
	"context"
	"errors"

	"github.com/redhat-developer/odo/pkg/api"

	bindingApi "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	specApi "github.com/redhat-developer/service-binding-operator/apis/spec/v1alpha3"

	ocappsv1 "github.com/openshift/api/apps/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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

var (
	NativeWorkloadKinds = []schema.GroupVersionKind{
		appsv1.SchemeGroupVersion.WithKind("DaemonSet"),
		appsv1.SchemeGroupVersion.WithKind("Deployment"),
		appsv1.SchemeGroupVersion.WithKind("ReplicaSet"),
		corev1.SchemeGroupVersion.WithKind("ReplicationController"),
		appsv1.SchemeGroupVersion.WithKind("StatefulSet"),
	}

	CustomWorkloadKinds = []schema.GroupVersionKind{
		ocappsv1.SchemeGroupVersion.WithKind("DeploymentConfig"),
	}
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
			return bindableKind, errors.New("No bindable operators found on the cluster. Please ensure that at least one bindable operator is installed successfully before proceeding. Known Bindable operators: https://github.com/redhat-developer/service-binding-operator#known-bindable-operators")
			//revive:enable:error-strings
		}
		return bindableKind, err
	}

	err = ConvertUnstructuredToResource(*unstructuredBK, &bindableKind)
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
func (c *Client) NewServiceBindingServiceObject(serviceNs string, unstructuredService unstructured.Unstructured, bindingName string) (bindingApi.Service, error) {
	serviceRESTMapping, err := c.GetRestMappingFromUnstructured(unstructuredService)
	if err != nil {
		return bindingApi.Service{}, err
	}

	var ns *string
	if serviceNs != "" {
		ns = &serviceNs
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
			Namespace: ns,
		},
	}, nil
}

// NewServiceBindingObject returns the bindingApi.ServiceBinding object
func NewServiceBindingObject(
	bindingName string,
	bindAsFiles bool,
	workloadName string,
	namingStrategy string,
	workloadGVK schema.GroupVersionKind,
	mappings []bindingApi.Mapping,
	services []bindingApi.Service,
	status bindingApi.ServiceBindingStatus,
) *bindingApi.ServiceBinding {
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
			NamingStrategy:         namingStrategy,
			Application: bindingApi.Application{
				Ref: bindingApi.Ref{
					Name:    workloadName,
					Group:   workloadGVK.Group,
					Version: workloadGVK.Version,
					Kind:    workloadGVK.Kind,
				},
			},
			Mappings: mappings,
			Services: services,
		},
		Status: status,
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
	err = ConvertUnstructuredToResource(*u, &result)
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
	err = ConvertUnstructuredToResource(*u, &result)
	if err != nil {
		return specApi.ServiceBinding{}, err
	}
	return result, nil
}

// ListServiceBindingsFromAllGroups returns the list of ServiceBindings in the cluster
// in the current namespace.
// The first list on the result contains ServiceBinding resources from group servicebinding.io/v1alpha3
// the second list contains ServiceBinding resources from group binding.operators.coreos.com/v1alpha1
func (c Client) ListServiceBindingsFromAllGroups() ([]specApi.ServiceBinding, []bindingApi.ServiceBinding, error) {
	if c.DynamicClient == nil {
		return nil, nil, nil
	}

	specsU, err := c.ListDynamicResources("", specApi.GroupVersionResource)
	var specs specApi.ServiceBindingList
	if err != nil {
		if !kerrors.IsForbidden(err) {
			return nil, nil, err
		}
	} else {
		err = ConvertUnstructuredListToResource(*specsU, &specs)
		if err != nil {
			return nil, nil, err
		}
	}

	bindingsU, err := c.ListDynamicResources("", bindingApi.GroupVersionResource)
	var bindings bindingApi.ServiceBindingList
	if err != nil {
		if !kerrors.IsForbidden(err) {
			return nil, nil, err
		}
	} else {
		err = ConvertUnstructuredListToResource(*bindingsU, &bindings)
		if err != nil {
			return nil, nil, err
		}
	}

	return specs.Items, bindings.Items, nil
}

// APIServiceBindingFromBinding returns a common api.ServiceBinding structure
// from a ServiceBinding.binding.operators.coreos.com/v1alpha1
func (c Client) APIServiceBindingFromBinding(binding bindingApi.ServiceBinding) (api.ServiceBinding, error) {

	var dstSvcs []corev1.ObjectReference
	for _, srcSvc := range binding.Spec.Services {
		dstSvc := corev1.ObjectReference{
			Name: srcSvc.Name,
		}
		dstSvc.APIVersion, dstSvc.Kind = schema.GroupVersion{
			Group:   srcSvc.Group,
			Version: srcSvc.Version,
		}.WithKind(srcSvc.Kind).ToAPIVersionAndKind()
		if srcSvc.Namespace != nil {
			dstSvc.Namespace = *srcSvc.Namespace
		}
		dstSvcs = append(dstSvcs, dstSvc)
	}

	application := binding.Spec.Application
	refToApplication := corev1.ObjectReference{
		Name: application.Name,
	}

	if application.Kind == "" {
		gvk, err := c.GetGVKFromGVR(schema.GroupVersionResource{
			Group:    application.Group,
			Version:  application.Version,
			Resource: application.Resource,
		})
		if err != nil {
			return api.ServiceBinding{}, err
		}
		application.Kind = gvk.Kind
	}
	refToApplication.APIVersion, refToApplication.Kind = schema.GroupVersion{
		Group:   application.Group,
		Version: application.Version,
	}.WithKind(application.Kind).ToAPIVersionAndKind()

	return api.ServiceBinding{
		Name: binding.Name,
		Spec: api.ServiceBindingSpec{
			Application:            refToApplication,
			Services:               dstSvcs,
			DetectBindingResources: binding.Spec.DetectBindingResources,
			BindAsFiles:            binding.Spec.BindAsFiles,
			NamingStrategy:         binding.Spec.NamingStrategy,
		},
	}, nil
}

// APIServiceBindingFromSpec returns a common api.ServiceBinding structure
// from a ServiceBinding.servicebinding.io/v1alpha3
func (c Client) APIServiceBindingFromSpec(spec specApi.ServiceBinding) api.ServiceBinding {

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

	return api.ServiceBinding{
		Name: spec.Name,
		Spec: api.ServiceBindingSpec{
			Application:            refToApplication,
			Services:               []corev1.ObjectReference{refToService},
			DetectBindingResources: false,
			BindAsFiles:            true,
		},
	}
}

// GetWorkloadKinds returns all the workload kinds present in the cluster
// It considers that all native resources are present and tests only for custom resources
// Returns an array of Kinds and an array of GVKs
func (c Client) GetWorkloadKinds() ([]string, []schema.GroupVersionKind, error) {
	var allWorkloadsKinds = []schema.GroupVersionKind{}
	var options []string
	for _, gvk := range NativeWorkloadKinds {
		options = append(options, gvk.Kind)
		allWorkloadsKinds = append(allWorkloadsKinds, gvk)
	}

	// Test for each custom workload kind if it exists in the cluster
	for _, gvk := range CustomWorkloadKinds {
		_, err := c.GetGVRFromGVK(gvk)
		if err != nil {
			// This is sufficient to test if resource exists in cluster
			if meta.IsNoMatchError(err) {
				continue
			}
			return nil, nil, err
		}
		options = append(options, gvk.Kind)
		allWorkloadsKinds = append(allWorkloadsKinds, gvk)
	}
	return options, allWorkloadsKinds, nil
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
