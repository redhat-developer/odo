package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DevWorkspaceTemplate is the Schema for the devworkspacetemplates API
// +kubebuilder:resource:path=devworkspacetemplates,scope=Namespaced,shortName=dwt
// +devfile:jsonschema:generate
// +kubebuilder:storageversion
type DevWorkspaceTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DevWorkspaceTemplateSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DevWorkspaceTemplateList contains a list of DevWorkspaceTemplate
type DevWorkspaceTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DevWorkspaceTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DevWorkspaceTemplate{}, &DevWorkspaceTemplateList{})
}
