package api

// ServiceBinding describes a service binding, from group binding.operators.coreos.com/v1alpha1 or servicebinding.io/v1alpha3
type ServiceBinding struct {
	Name   string                `json:"name"`
	Spec   ServiceBindingSpec    `json:"spec"`
	Status *ServiceBindingStatus `json:"status,omitempty"`
}

type ServiceBindingSpec struct {
	Application            ServiceBindingReference   `json:"application"`
	Services               []ServiceBindingReference `json:"services"`
	DetectBindingResources bool                      `json:"detectBindingResources"`
	BindAsFiles            bool                      `json:"bindAsFiles"`
	NamingStrategy         string                    `json:"namingStrategy,omitempty"`
}

type ServiceBindingReference struct {
	Kind       string `json:"kind,omitempty"`
	Resource   string `json:"resource,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	Name       string `json:"name,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
}

type ServiceBindingStatus struct {
	BindingFiles   []string     `json:"bindingFiles,omitempty"`
	BindingEnvVars []string     `json:"bindingEnvVars,omitempty"`
	RunningIn      RunningModes `json:"runningIn,omitempty"`
}
