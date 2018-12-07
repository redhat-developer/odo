package component

import (
	"github.com/redhat-developer/odo/pkg/storage"
	urlpkg "github.com/redhat-developer/odo/pkg/url"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// machine readable struct
type Component struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ComponentSpec `json:"spec,omitempty"`
}

// ComponentList is a list of Components.
type ComponentList struct {
	metav1.TypeMeta `json:",inline"`

	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []CompoList `json:"items"`
}

// ComponentSpec holds all information about component
type ComponentSpec struct {
	ComponentName      string            `json:"name,omitempty"`
	ComponentImageType string            `json:"type,omitempty"`
	Path               string            `json:"source,omitempty"`
	URLs               []urlpkg.UrlSpec  `json:"url,omitempty"`
	Env                []corev1.EnvVar   `json:"environment,omitempty"`
	Storage            []storage.Storage `json:"storage,omitempty"`
}

type CompoList struct {
	Name   string
	Type   string
	Active bool
}
