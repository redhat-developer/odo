package component

import (
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/occlient"
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
	Type    string          `json:"type,omitempty"`
	Source  string          `json:"source,omitempty"`
	URL     []string        `json:"url,omitempty"`
	Storage []string        `json:"storage,omitempty"`
	Env     []corev1.EnvVar `json:"env,omitempty"`
}

// ComponentList is list of components
type ComponentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Component `json:"items"`
}

// ComponentStatus is Status of components
type ComponentStatus struct {
	Active           bool                `json:"active"`
	LinkedComponents map[string][]string `json:"linkedComponents,omitempty"`
	LinkedServices   []string            `json:"linkedServices,omitempty"`
}

// ComponentOptions are generic options that are used between commands such as watch and push
type ComponentOptions struct {
	SourceType       config.SrcType
	SourcePath       string
	LocalConfig      *config.LocalConfigInfo
	ComponentContext string
	Client           *occlient.Client
}
