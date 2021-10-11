package v1alpha2

import (
	attributes "github.com/devfile/api/v2/pkg/attributes"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// ComponentType describes the type of component.
// Only one of the following component type may be specified.
// +kubebuilder:validation:Enum=Container;Kubernetes;Openshift;Volume;Image;Plugin;Custom
type ComponentType string

const (
	ContainerComponentType  ComponentType = "Container"
	KubernetesComponentType ComponentType = "Kubernetes"
	OpenshiftComponentType  ComponentType = "Openshift"
	PluginComponentType     ComponentType = "Plugin"
	VolumeComponentType     ComponentType = "Volume"
	ImageComponentType      ComponentType = "Image"
	CustomComponentType     ComponentType = "Custom"
)

// DevWorkspace component: Anything that will bring additional features / tooling / behaviour / context
// to the devworkspace, in order to make working in it easier.
type BaseComponent struct {
}

//+k8s:openapi-gen=true
type Component struct {
	// Mandatory name that allows referencing the component
	// from other elements (such as commands) or from an external
	// devfile that may reference this component through a parent or a plugin.
	// +kubebuilder:validation:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	// +kubebuilder:validation:MaxLength=63
	Name string `json:"name"`
	// Map of implementation-dependant free-form YAML attributes.
	// +optional
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Attributes     attributes.Attributes `json:"attributes,omitempty"`
	ComponentUnion `json:",inline"`
}

// +union
type ComponentUnion struct {
	// Type of component
	//
	// +unionDiscriminator
	// +optional
	ComponentType ComponentType `json:"componentType,omitempty"`

	// Allows adding and configuring devworkspace-related containers
	// +optional
	Container *ContainerComponent `json:"container,omitempty"`

	// Allows importing into the devworkspace the Kubernetes resources
	// defined in a given manifest. For example this allows reusing the Kubernetes
	// definitions used to deploy some runtime components in production.
	//
	// +optional
	Kubernetes *KubernetesComponent `json:"kubernetes,omitempty"`

	// Allows importing into the devworkspace the OpenShift resources
	// defined in a given manifest. For example this allows reusing the OpenShift
	// definitions used to deploy some runtime components in production.
	//
	// +optional
	Openshift *OpenshiftComponent `json:"openshift,omitempty"`

	// Allows specifying the definition of a volume
	// shared by several other components
	// +optional
	Volume *VolumeComponent `json:"volume,omitempty"`

	// Allows specifying the definition of an image for outer loop builds
	// +optional
	Image *ImageComponent `json:"image,omitempty"`

	// Allows importing a plugin.
	//
	// Plugins are mainly imported devfiles that contribute components, commands
	// and events as a consistent single unit. They are defined in either YAML files
	// following the devfile syntax,
	// or as `DevWorkspaceTemplate` Kubernetes Custom Resources
	// +optional
	// +devfile:overrides:include:omitInPlugin=true
	Plugin *PluginComponent `json:"plugin,omitempty"`

	// Custom component whose logic is implementation-dependant
	// and should be provided by the user
	// possibly through some dedicated controller
	// +optional
	// +devfile:overrides:include:omit=true
	Custom *CustomComponent `json:"custom,omitempty"`
}

type CustomComponent struct {
	// Class of component that the associated implementation controller
	// should use to process this command with the appropriate logic
	ComponentClass string `json:"componentClass"`

	// Additional free-form configuration for this custom component
	// that the implementation controller will know how to use
	//
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:EmbeddedResource
	EmbeddedResource runtime.RawExtension `json:"embeddedResource"`
}
