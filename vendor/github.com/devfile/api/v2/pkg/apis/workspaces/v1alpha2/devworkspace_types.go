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
	// Id of the workspace
	WorkspaceId string `json:"workspaceId"`
	// URL at which the Worksace Editor can be joined
	IdeUrl string         `json:"ideUrl,omitempty"`
	Phase  WorkspacePhase `json:"phase,omitempty"`
	// Conditions represent the latest available observations of an object's state
	Conditions []WorkspaceCondition `json:"conditions,omitempty"`
	// Message is a short user-readable message giving additional information
	// about an object's state
	Message string `json:"message,omitempty"`
}

type WorkspacePhase string

// Valid workspace Statuses
const (
	WorkspaceStatusStarting WorkspacePhase = "Starting"
	WorkspaceStatusRunning  WorkspacePhase = "Running"
	WorkspaceStatusStopped  WorkspacePhase = "Stopped"
	WorkspaceStatusStopping WorkspacePhase = "Stopping"
	WorkspaceStatusFailed   WorkspacePhase = "Failed"
	WorkspaceStatusError    WorkspacePhase = "Error"
)

// WorkspaceCondition contains details for the current condition of this workspace.
type WorkspaceCondition struct {
	// Type is the type of the condition.
	Type WorkspaceConditionType `json:"type"`
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

// Types of conditions reported by workspace
type WorkspaceConditionType string

const (
	WorkspaceComponentsReady     WorkspaceConditionType = "ComponentsReady"
	WorkspaceRoutingReady        WorkspaceConditionType = "RoutingReady"
	WorkspaceServiceAccountReady WorkspaceConditionType = "ServiceAccountReady"
	WorkspaceReady               WorkspaceConditionType = "Ready"
	WorkspaceFailedStart         WorkspaceConditionType = "FailedStart"
	WorkspaceError               WorkspaceConditionType = "Error"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DevWorkspace is the Schema for the devworkspaces API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=devworkspaces,scope=Namespaced,shortName=dw
// +kubebuilder:printcolumn:name="Workspace ID",type="string",JSONPath=".status.workspaceId",description="The workspace's unique id"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="The current workspace startup phase"
// +kubebuilder:printcolumn:name="Info",type="string",JSONPath=".status.message",description="Additional information about the workspace"
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
