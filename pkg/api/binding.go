package api

import (
	corev1 "k8s.io/api/core/v1"
)

// ServiceBinding describes a service binding, from group binding.operators.coreos.com/v1alpha1 or servicebinding.io/v1alpha3
type ServiceBinding struct {
	Name   string                `json:"name"`
	Spec   ServiceBindingSpec    `json:"spec"`
	Status *ServiceBindingStatus `json:"status,omitempty"`
}

type ServiceBindingSpec struct {
	Application            corev1.ObjectReference   `json:"application"`
	Services               []corev1.ObjectReference `json:"services"`
	DetectBindingResources bool                     `json:"detectBindingResources"`
	BindAsFiles            bool                     `json:"bindAsFiles"`
	NamingStrategy         string                   `json:"namingStrategy,omitempty"`
}

type ServiceBindingStatus struct {
	BindingFiles   []string        `json:"bindingFiles,omitempty"`
	BindingEnvVars []string        `json:"bindingEnvVars,omitempty"`
	RunningIn      RunningModeList `json:"runningIn,omitempty"`
}
