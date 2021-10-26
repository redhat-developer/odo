package component

import (
	"github.com/openshift/odo/v2/pkg/machineoutput"
	"github.com/openshift/odo/v2/pkg/storage"
	"github.com/openshift/odo/v2/pkg/url"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ComponentKind = "Component"

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
	URL         []string          `json:"url,omitempty"`
	URLSpec     []url.URL         `json:"-"`
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
	State          State         `json:"state"`
	LinkedServices []SecretMount `json:"linkedServices,omitempty"`
}

// CombinedComponentList is list of s2i and devfile components
type CombinedComponentList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ListMeta   `json:"metadata,omitempty"`
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

func newComponentWithType(componentName, componentType string) Component {
	cmp := NewComponent(componentName)
	cmp.Spec.Type = componentType
	return cmp
}

// NewComponent provides a constructor to component struct with some metadata prefilled
func NewComponent(componentName string) Component {
	return Component{
		TypeMeta: metav1.TypeMeta{
			Kind:       ComponentKind,
			APIVersion: machineoutput.APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: componentName,
		},
		Status: ComponentStatus{},
	}
}

// newComponentList returns list of devfile and s2i components in machine readable format
func newComponentList(comps []Component) ComponentList {
	if len(comps) == 0 {
		comps = []Component{}
	}

	return ComponentList{
		TypeMeta: metav1.TypeMeta{
			Kind:       machineoutput.ListKind,
			APIVersion: machineoutput.APIVersion,
		},
		ListMeta: metav1.ListMeta{},
		Items:    comps,
	}
}

// NewCombinedComponentList returns list of devfile, s2i components and other components(not managed by odo) in machine readable format
func NewCombinedComponentList(devfileComps []Component, otherComps []Component) CombinedComponentList {

	if len(devfileComps) == 0 {
		devfileComps = []Component{}
	}
	if len(otherComps) == 0 {
		otherComps = []Component{}
	}

	return CombinedComponentList{
		TypeMeta: metav1.TypeMeta{
			Kind:       machineoutput.ListKind,
			APIVersion: machineoutput.APIVersion,
		},
		ListMeta:          metav1.ListMeta{},
		DevfileComponents: devfileComps,
		OtherComponents:   otherComps,
	}
}
