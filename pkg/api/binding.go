package api

import (
	sbcApi "github.com/redhat-developer/service-binding-operator/apis/spec/v1alpha3"
)

type ServiceBinding struct {
	Name   string                `json:"name"`
	Spec   ServiceBindingSpec    `json:"spec"`
	Status *ServiceBindingStatus `json:"status,omitempty"`
}

type ServiceBindingSpec struct {
	Services               []sbcApi.ServiceBindingServiceReference `json:"services"`
	DetectBindingResources bool                                    `json:"detectBindingResources"`
	BindAsFiles            bool                                    `json:"bindAsFiles"`
}

type ServiceBindingStatus struct {
	BindingFiles   []string `json:"bindingsFiles,omitempty"`
	BindingEnvVars []string `json:"bindingEnvVars,omitempty"`
}
