package service

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type OperatorExample struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              map[string]interface{} `json:"spec,omitempty"`
}

// ServiceInfo holds all important information about one service
type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ServiceSpec   `json:"spec,omitempty"`
	Status            ServiceStatus `json:"status,omitempty"`
}

// ServiceSpec ...
type ServiceSpec struct {
	Type string `json:"type,omitempty"`
	Plan string `json:"plan,omitempty"`
}

// ServiceStatus ...
type ServiceStatus struct {
	Status string `json:"status,omitempty"`
}

type ServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Service `json:"items"`
}

func NewOperatorExample(almExample map[string]interface{}) OperatorExample {
	return OperatorExample{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OperatorExample",
			APIVersion: "odo.dev/v1alpha1",
		},
		Spec: almExample,
	}
}
