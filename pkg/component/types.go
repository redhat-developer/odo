package component

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redhat-developer/odo/pkg/storage"
)

const ComponentKind = "Component"

// Component
type Component struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ComponentSpec   `json:"spec,omitempty"`
	Status            ComponentStatus `json:"status,omitempty"`
}

// OdoComponent
type OdoComponent struct {
	Name      string
	ManagedBy string
	Modes     map[string]bool
	Type      string
}

// ComponentSpec is spec of components
type ComponentSpec struct {
	App         string            `json:"app,omitempty"`
	Type        string            `json:"type,omitempty"`
	Managed     string            `json:"managed,omitempty"`
	Dev         bool              `json:"dev,omitempty"`
	Deploy      bool              `json:"deploy,omitempty"`
	Source      string            `json:"source,omitempty"`
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

// SecretMount describes a Secret mount (either as environment variables with envFrom or as a volume)
type SecretMount struct {
	ServiceName string
	SecretName  string
	MountVolume bool
	MountPath   string
}

// ComponentStatus is Status of components
type ComponentStatus struct {
	Context        string        `json:"context,omitempty"`
	State          string        `json:"state"`
	LinkedServices []SecretMount `json:"linkedServices,omitempty"`
}

// CombinedComponentList is list of s2i and devfile components
type CombinedComponentList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ListMeta   `json:"metadata,omitempty"`
	DevfileComponents []Component `json:"devfileComponents"`
	OtherComponents   []Component `json:"otherComponents"`
}

const (
	// StateTypeUnknown means that odo cannot tell its state
	StateTypeUnknown = "Unknown"
	// StateTypeNone means that it has not been pushed to the cluster *at all* in either deploy or dev
	StateTypeNone = "None"
)
