package catalog

import (
	imagev1 "github.com/openshift/api/image/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ComponentType is the main struct for catalog components
type ComponentType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ComponentSpec `json:"spec,omitempty"`
}

// ComponentSpec is the spec for ComponentType
type ComponentSpec struct {
	AllTags        []string            `json:"allTags"`
	NonHiddenTags  []string            `json:"nonHiddenTags"`
	ImageStreamRef imagev1.ImageStream `json:"imageStreamRef"`
}

// ComponentTypeList lists all the ComponentType's
type ComponentTypeList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Items             []ComponentType `json:"items"`
}

// ServiceType is the main struct for catalog services
type ServiceType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ServiceSpec `json:"spec,omitempty"`
}

// ServiceSpec is the spec for ServiceType
type ServiceSpec struct {
	Hidden   bool     `json:"hidden"`
	PlanList []string `json:"planList"`
}

// ServiceTypeList lists all the ServiceType's
type ServiceTypeList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Items             []ServiceType `json:"items"`
}
