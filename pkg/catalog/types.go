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

// DevfileComponentType is the main struct for devfile catalog components
type DevfileComponentType struct {
	Name        string
	DisplayName string
	Description string
	Link        string
	Support     bool
	Registry    string
}

// DevfileIndexEntry is the main struct of index.json from devfile registry
type DevfileIndexEntry struct {
	DisplayName       string   `json:"displayName"`
	Description       string   `json:"description"`
	Tags              []string `json:"tags"`
	Icon              string   `json:"icon"`
	GlobalMemoryLimit string   `json:"globalMemoryLimit"`
	Links             struct {
		Link string `json:"self"`
	} `json:"links"`
}

// Devfile is the main structure of devfile from devfile registry
type Devfile struct {
	APIVersion string `yaml:"apiVersion"`
	MetaData   struct {
		GenerateName string `yaml:"generateName"`
	} `yaml:"metadata"`
	Components []struct {
		Type  string `yaml:"type"`
		Alias string `yaml:"alias"`
	} `yaml:"components"`
	Commands []struct {
		Name string `yaml:"name"`
	} `yaml:"commands"`
}

// ComponentSpec is the spec for ComponentType
type ComponentSpec struct {
	AllTags        []string            `json:"allTags"`
	NonHiddenTags  []string            `json:"nonHiddenTags"`
	SupportedTags  []string            `json:"supportedTags"`
	ImageStreamRef imagev1.ImageStream `json:"-"`
}

// ComponentTypeList lists all the ComponentType's
type ComponentTypeList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Items             []ComponentType `json:"items"`
}

// DevfileComponentTypeList lists all the DevfileComponentType's
type DevfileComponentTypeList struct {
	DevfileRegistries []string
	Items             []DevfileComponentType
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
