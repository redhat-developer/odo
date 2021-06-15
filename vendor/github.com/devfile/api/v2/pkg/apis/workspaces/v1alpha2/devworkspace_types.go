package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DevWorkspaceSpec defines the desired state of DevWorkspace
type DevWorkspaceSpec struct {
	Started      bool                     `json:"started"`
	RoutingClass string                   `json:"routingClass,omitempty"`
	Template     DevWorkspaceTemplateSpec `json:"template,omitempty"`
}

// DevWorkspaceStatus defines the observed state of DevWorkspace
type DevWorkspaceStatus struct {
	// Id of the DevWorkspace
	DevWorkspaceId string `json:"devworkspaceId"`
	// Main URL for this DevWorkspace
	MainUrl string            `json:"mainUrl,omitempty"`
	Phase   DevWorkspacePhase `json:"phase,omitempty"`
	// Conditions represent the latest available observations of an object's state
	Conditions []DevWorkspaceCondition `json:"conditions,omitempty"`
	// Message is a short user-readable message giving additional information
	// about an object's state
	Message string `json:"message,omitempty"`
}

type DevWorkspacePhase string

// Valid devworkspace Statuses
const (
	DevWorkspaceStatusStarting DevWorkspacePhase = "Starting"
	DevWorkspaceStatusRunning  DevWorkspacePhase = "Running"
	DevWorkspaceStatusStopped  DevWorkspacePhase = "Stopped"
	DevWorkspaceStatusStopping DevWorkspacePhase = "Stopping"
	DevWorkspaceStatusFailed   DevWorkspacePhase = "Failed"
	DevWorkspaceStatusError    DevWorkspacePhase = "Error"
)

// DevWorkspaceCondition contains details for the current condition of this devworkspace.
type DevWorkspaceCondition struct {
	// Type is the type of the condition.
	Type DevWorkspaceConditionType `json:"type"`
	// Phase is the status of the condition.
	// Can be True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	Message string `json:"message,omitempty"`
}

// Types of conditions reported by devworkspace
type DevWorkspaceConditionType string

const (
	DevWorkspaceComponentsReady     DevWorkspaceConditionType = "ComponentsReady"
	DevWorkspaceRoutingReady        DevWorkspaceConditionType = "RoutingReady"
	DevWorkspaceServiceAccountReady DevWorkspaceConditionType = "ServiceAccountReady"
	DevWorkspaceReady               DevWorkspaceConditionType = "Ready"
	DevWorkspaceFailedStart         DevWorkspaceConditionType = "FailedStart"
	DevWorkspaceError               DevWorkspaceConditionType = "Error"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DevWorkspace is the Schema for the devworkspaces API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=devworkspaces,scope=Namespaced,shortName=dw
// +kubebuilder:printcolumn:name="DevWorkspace ID",type="string",JSONPath=".status.devworkspaceId",description="The devworkspace's unique id"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="The current devworkspace startup phase"
// +kubebuilder:printcolumn:name="Info",type="string",JSONPath=".status.message",description="Additional information about the devworkspace"
// +devfile:jsonschema:generate
// +kubebuilder:storageversion
type DevWorkspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DevWorkspaceSpec   `json:"spec,omitempty"`
	Status DevWorkspaceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DevWorkspaceList contains a list of DevWorkspace
type DevWorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DevWorkspace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DevWorkspace{}, &DevWorkspaceList{})
}
