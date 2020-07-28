package servicebindingrequest

import (
	"github.com/imdario/mergo"
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/annotations"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"

	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

// ServiceContext contains information related to a service.
type ServiceContext struct {
	// Service is the resource of the service being evaluated.
	Service *unstructured.Unstructured
	// EnvVars contains the service's contributed environment variables.
	EnvVars map[string]interface{}
	// VolumeKeys contains the keys that should be mounted as volume from the binding secret.
	VolumeKeys []string
	// EnvVarPrefix indicates the prefix to use in environment variables.
	EnvVarPrefix *string
}

// ServiceContextList is a list of ServiceContext values.
type ServiceContextList []*ServiceContext

// GetServices returns a slice of service unstructured objects contained in the collection.
func (sc ServiceContextList) GetServices() []*unstructured.Unstructured {
	var crs []*unstructured.Unstructured
	for _, s := range sc {
		crs = append(crs, s.Service)
	}
	return crs
}

func stringValueOrDefault(val *string, defaultVal string) string {
	if val != nil && len(*val) > 0 {
		return *val
	}
	return defaultVal
}

// buildServiceContexts return a collection of ServiceContext values from the given service
// selectors.
func buildServiceContexts(
	client dynamic.Interface,
	defaultNs string,
	selectors []v1alpha1.BackingServiceSelector,
	includeServiceOwnedResources bool,
	restMapper meta.RESTMapper,
) (ServiceContextList, error) {
	svcCtxs := make(ServiceContextList, 0)
	for _, s := range selectors {
		ns := stringValueOrDefault(s.Namespace, defaultNs)
		gvk := schema.GroupVersionKind{Kind: s.Kind, Version: s.Version, Group: s.Group}
		svcCtx, err := buildServiceContext(
			client, ns, gvk, s.ResourceRef, s.EnvVarPrefix, restMapper)
		if err != nil {
			return nil, err
		}
		svcCtxs = append(svcCtxs, svcCtx)

		if includeServiceOwnedResources {
			// use the selector's kind as owned resources environment variable prefix
			svcEnvVarPrefix := svcCtx.EnvVarPrefix
			if svcEnvVarPrefix == nil {
				svcEnvVarPrefix = &s.Kind
			}
			ownedResourcesCtxs, err := findOwnedResourcesCtxs(
				client,
				ns,
				svcCtx.Service.GetName(),
				svcCtx.Service.GetUID(),
				gvk,
				svcEnvVarPrefix,
				restMapper,
			)
			if err != nil {
				return nil, err
			}
			svcCtxs = append(svcCtxs, ownedResourcesCtxs...)
		}
	}

	return svcCtxs, nil
}

func findOwnedResourcesCtxs(
	client dynamic.Interface,
	ns string,
	name string,
	uid types.UID,
	gvk schema.GroupVersionKind,
	envVarPrefix *string,
	restMapper meta.RESTMapper,
) (ServiceContextList, error) {
	ownedResources, err := getOwnedResources(
		client,
		ns,
		gvk,
		name,
		uid,
	)
	if err != nil {
		return nil, err
	}

	return buildOwnedResourceContexts(
		client,
		ownedResources,
		envVarPrefix,
		restMapper,
	)
}

// buildServiceContext inspects g the API server searching for the service resources, associated CRD
// and OLM's CRDDescription if present, and processes those with relevant annotations to compose a
// ServiceContext.
func buildServiceContext(
	client dynamic.Interface,
	ns string,
	gvk schema.GroupVersionKind,
	resourceRef string,
	envVarPrefix *string,
	restMapper meta.RESTMapper,
) (*ServiceContext, error) {
	obj, err := findService(client, ns, gvk, resourceRef)
	if err != nil {
		return nil, err
	}

	anns := map[string]string{}

	// attempt to search the CRD of given gvk and bail out right away if a CRD can't be found; this
	// means also a CRDDescription can't exist or if it does exist it is not meaningful.
	crd, err := findServiceCRD(client, gvk)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	} else if !errors.IsNotFound(err) {
		// attempt to search the a CRDDescription related to the obtained CRD.
		crdDescription, err := findCRDDescription(ns, client, gvk, crd)
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		}
		// start with annotations extracted from CRDDescription
		err = mergo.Merge(
			&anns, convertCRDDescriptionToAnnotations(crdDescription), mergo.WithOverride)
		if err != nil {
			return nil, err
		}
		// then override collected annotations with CRD annotations
		err = mergo.Merge(&anns, crd.GetAnnotations(), mergo.WithOverride)
		if err != nil {
			return nil, err
		}
	}

	// and finally override collected annotations with own annotations
	err = mergo.Merge(&anns, obj.GetAnnotations(), mergo.WithOverride)
	if err != nil {
		return nil, err
	}

	volumeKeys := make([]string, 0)
	envVars := make(map[string]interface{})

	for annotationKey, annotationValue := range anns {
		h, err := annotations.BuildHandler(
			client,
			obj,
			annotationKey,
			annotationValue,
			restMapper,
		)
		if err != nil {
			if err == annotations.ErrInvalidAnnotationPrefix || annotations.IsErrHandlerNotFound(err) {
				continue
			}
			return nil, err
		}
		r, err := h.Handle()
		if err != nil {
			continue
		}

		err = mergo.Merge(&envVars, r.Data, mergo.WithAppendSlice, mergo.WithOverride)
		if err != nil {
			return nil, err
		}

		if r.Type == annotations.BindingTypeVolumeMount {
			volumeKeys = append(volumeKeys, r.Path)
		}
	}

	serviceCtx := &ServiceContext{
		Service:      obj,
		EnvVars:      envVars,
		VolumeKeys:   volumeKeys,
		EnvVarPrefix: envVarPrefix,
	}

	return serviceCtx, nil
}
