package service

//OperatorBackend implements the interface ServiceProviderBackend and contains methods that help create a service from Operators
type OperatorBackend struct {
	// Custom Resrouce to create service from
	CustomResource string
	// Custom Resrouce's Definition fetched from alm-examples
	CustomResourceDefinition map[string]interface{}
	// Group of the GVR
	group string
	// Version of the GVR
	version string
	// Resource of the GVR
	resource string
	// Kind of GVK
	kind string
}

func NewOperatorBackend() *OperatorBackend {
	return &OperatorBackend{}
}

// ServiceCatalogBackend implements the interface ServiceProviderBackend and contains methods that help create a service from Service Catalog
type ServiceCatalogBackend struct {
}

func NewServiceCatalogBackend() *ServiceCatalogBackend {
	return &ServiceCatalogBackend{}
}
