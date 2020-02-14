package clientset

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Clientset maps GroupVersionResources to underlying dynamic clients. If the
// GVR does not exist, operations will return an error.
type Clientset struct {
	config map[schema.GroupVersionResource]dynamic.Interface
}

// Option defines optional configuration for the Clientset. Most commonly used
// to initialize extensions.
type Option func(*Clientset)

// New creates a new Clientset with the provided options.
func New(opts ...Option) *Clientset {
	cs := &Clientset{
		config: make(map[schema.GroupVersionResource]dynamic.Interface),
	}
	for _, o := range opts {
		o(cs)
	}

	return cs
}

// Add adds a new mapping for the given resource.
func (r *Clientset) Add(resource schema.GroupVersionResource, client dynamic.Interface) {
	r.config[resource] = client
}

// Resource returns the dynamic Resource for the given GVR. If not configured,
// an error resource is returned.
func (r *Clientset) Resource(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	i, ok := r.config[resource]
	if !ok {
		return newErrorResource(resource)
	}
	return i.Resource(resource)
}
