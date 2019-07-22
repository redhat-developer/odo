package service

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Service represents service object
type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ServiceSpec
	Status            ServiceStatus
}

// ServiceSpec is spec for service object
type ServiceSpec struct {
	ServiceType   string            `json:"type,omitempty"`
	ServicePlan   string            `json:"plan,omitempty"`
	ParametersMap map[string]string `json:"parameters,omitempty"`
}

// ServiceStatus is status for service object
type ServiceStatus struct {
	Message string `json:"message,omitempty"`
}

type ServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Service `json:"items"`
}
