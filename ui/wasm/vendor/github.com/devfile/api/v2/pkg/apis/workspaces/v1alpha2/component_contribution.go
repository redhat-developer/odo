package v1alpha2

import attributes "github.com/devfile/api/v2/pkg/attributes"

type ComponentContribution struct {
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
	Attributes attributes.Attributes `json:"attributes,omitempty"`
	// Reference to a remote Devfile or DevWorkspace resource to be included
	// in the DevWorkspace.
	PluginComponent `json:",inline"`
}
