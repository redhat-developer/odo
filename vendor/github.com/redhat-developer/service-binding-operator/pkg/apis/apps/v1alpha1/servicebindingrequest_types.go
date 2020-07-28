package v1alpha1

import (
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

// ServiceBindingRequestSpec defines the desired state of ServiceBindingRequest
type ServiceBindingRequestSpec struct {
	// MountPathPrefix is the prefix for volume mount
	// +optional
	MountPathPrefix string `json:"mountPathPrefix,omitempty"`

	// EnvVarPrefix is the prefix for environment variables
	// +optional
	EnvVarPrefix string `json:"envVarPrefix,omitempty"`

	// Custom env variables
	// +optional
	CustomEnvVar []corev1.EnvVar `json:"customEnvVar,omitempty"`

	// BackingServiceSelector is used to identify the backing service operator.
	// Deprecation Notice:
	// In the upcoming release, this field would be depcreated. It would be mandatory
	// to set "backingServiceSelectors".
	// +optional
	BackingServiceSelector *BackingServiceSelector `json:"backingServiceSelector,omitempty"`

	// BackingServiceSelectors is used to identify multiple backing services.
	// This would be made a required field after 'BackingServiceSelector'
	// is removed.
	// +optional
	BackingServiceSelectors *[]BackingServiceSelector `json:"backingServiceSelectors,omitempty"`

	// ApplicationSelector is used to identify the application connecting to the
	// backing service operator.
	// +optional
	ApplicationSelector ApplicationSelector `json:"applicationSelector"`

	// DetectBindingResources is flag used to bind all non-bindable variables from
	// different subresources owned by backing operator CR.
	// +optional
	DetectBindingResources bool `json:"detectBindingResources,omitempty"`
}

// ServiceBindingRequestStatus defines the observed state of ServiceBindingRequest
// +k8s:openapi-gen=true
type ServiceBindingRequestStatus struct {
	// Conditions describes the state of the operator's reconciliation functionality.
	Conditions []conditionsv1.Condition `json:"conditions"`
	// Secret is the name of the intermediate secret
	Secret string `json:"secret"`
	// Applications contain all the applications filtered by name or label
	// +optional
	Applications []BoundApplication `json:"applications,omitempty"`
}

// BackingServiceSelector defines the selector based on resource name, version, and resource kind
type BackingServiceSelector struct {
	metav1.GroupVersionKind `json:",inline"`
	ResourceRef             string `json:"resourceRef"`
	// +optional
	Namespace    *string `json:"namespace,omitempty"`
	EnvVarPrefix *string `json:"envVarPrefix,omitempty"`
}

// BoundApplication defines the application workloads to which the binding secret has
// injected.
type BoundApplication struct {
	metav1.GroupVersionKind     `json:",inline"`
	corev1.LocalObjectReference `json:",inline"`
}

// ApplicationSelector defines the selector based on labels and GVR
type ApplicationSelector struct {
	// +optional
	LabelSelector               *metav1.LabelSelector `json:"labelSelector,omitempty"`
	metav1.GroupVersionResource `json:",inline"`
	ResourceRef                 string `json:"resourceRef,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceBindingRequest expresses intent to bind an operator-backed service with
// an application workload.
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Service Binding Request"
// +kubebuilder:resource:path=servicebindingrequests,shortName=sbr;sbrs
type ServiceBindingRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceBindingRequestSpec   `json:"spec,omitempty"`
	Status ServiceBindingRequestStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceBindingRequestList contains a list of ServiceBindingRequest
type ServiceBindingRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ServiceBindingRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceBindingRequest{}, &ServiceBindingRequestList{})
}
