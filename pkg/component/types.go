package component

import (
	"github.com/openshift/odo/pkg/storage"
	"github.com/openshift/odo/pkg/urltype"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Component
type Component struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ComponentSpec   `json:"spec,omitempty"`
	Status            ComponentStatus `json:"status,omitempty"`
}

// ComponentSpec is spec of components
type ComponentSpec struct {
	App         string            `json:"app,omitempty"`
	Type        string            `json:"type,omitempty"`
	Source      string            `json:"source,omitempty"`
	SourceType  string            `json:"sourceType,omitempty"`
	URL         []string          `json:"url,omitempty"`
	URLSpec     []urltype.URL     `json:"-"`
	Storage     []string          `json:"storage,omitempty"`
	StorageSpec []storage.Storage `json:"-"`
	Env         []corev1.EnvVar   `json:"env,omitempty"`
	Ports       []string          `json:"ports,omitempty"`
}

// ComponentList is list of components
type ComponentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Component `json:"items"`
}

// ComponentStatus is Status of components
type ComponentStatus struct {
	Context          string              `json:"context,omitempty"`
	State            State               `json:"state"`
	LinkedComponents map[string][]string `json:"linkedComponents,omitempty"`
	LinkedServices   []string            `json:"linkedServices,omitempty"`
}

// CombinedComponentList is list of s2i and devfile components
type CombinedComponentList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ListMeta   `json:"metadata,omitempty"`
	S2IComponents     []Component `json:"s2iComponents"`
	DevfileComponents []Component `json:"devfileComponents"`
	OtherComponents   []Component `json:"otherComponents"`
}

// State represents the component state
type State string

const (
	// StateTypePushed means that Storage is present both locally and on cluster
	StateTypePushed State = "Pushed"
	// StateTypeNotPushed means that Storage is only in local config, but not on the cluster
	StateTypeNotPushed State = "Not Pushed"
	// StateTypeUnknown means that odo cannot tell its state
	StateTypeUnknown State = "Unknown"
)
