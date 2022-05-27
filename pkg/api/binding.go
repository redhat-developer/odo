package api

import (
	sboApi "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
)

//// SPEC
//type A1ServiceBindingSpec struct {
//	// Name is the name of the service as projected into the workload container.  Defaults to .metadata.name.
//	Name string `json:"name,omitempty"`
//	// Type is the type of the service as projected into the workload container
//	Type string `json:"type,omitempty"`
//	// Provider is the provider of the service as projected into the workload container
//	Provider string `json:"provider,omitempty"`
//	// Workload is a reference to an object
//	Workload ServiceBindingWorkloadReference `json:"workload"`
//	// Service is a reference to an object that fulfills the ProvisionedService duck type
//	Service ServiceBindingServiceReference `json:"service"`
//	// Env is the collection of mappings from Secret entries to environment variables
//	Env []EnvMapping `json:"env,omitempty"`
//}
//

type ServiceBinding struct {
	Name string `json:"name"`

	Services []sboApi.Service `json:"services"`

	Application sboApi.Application `json:"application"`

	DetectBindingResources bool `json:"detectBindingResources"`

	BindAsFiles bool `json:"bindAsFiles"`
}
