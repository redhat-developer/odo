package kubernetes

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//go:generate mockgen -destination=mocks/mocks.go -package=mocks . K8STypeLookup

type K8STypeLookup interface {
	ResourceForReferable(obj Referable) (*schema.GroupVersionResource, error)
	ResourceForKind(gvk schema.GroupVersionKind) (*schema.GroupVersionResource, error)
	KindForResource(gvr schema.GroupVersionResource) (*schema.GroupVersionKind, error)
}

type resourceLookup struct {
	restMapper meta.RESTMapper
}

func ResourceLookup(restMapper meta.RESTMapper) K8STypeLookup {
	return &resourceLookup{
		restMapper: restMapper,
	}
}

func (i *resourceLookup) ResourceForReferable(obj Referable) (*schema.GroupVersionResource, error) {
	gvr, err := obj.GroupVersionResource()
	if err == nil {
		return gvr, nil
	}
	gvk, err := obj.GroupVersionKind()
	if err != nil {
		return nil, err
	}
	return i.ResourceForKind(*gvk)
}

func (i *resourceLookup) ResourceForKind(gvk schema.GroupVersionKind) (*schema.GroupVersionResource, error) {
	mapping, err := i.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	return &mapping.Resource, nil
}

func (i *resourceLookup) KindForResource(gvr schema.GroupVersionResource) (*schema.GroupVersionKind, error) {
	gvk, err := i.restMapper.KindFor(gvr)
	return &gvk, err
}
