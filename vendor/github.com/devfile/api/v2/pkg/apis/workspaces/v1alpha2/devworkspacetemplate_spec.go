package v1alpha2

import attributes "github.com/devfile/api/v2/pkg/attributes"

// Structure of the devworkspace. This is also the specification of a devworkspace template.
// +devfile:jsonschema:generate
type DevWorkspaceTemplateSpec struct {
	// Parent devworkspace template
	// +optional
	Parent *Parent `json:"parent,omitempty"`

	DevWorkspaceTemplateSpecContent `json:",inline"`
}

// +devfile:overrides:generate
type DevWorkspaceTemplateSpecContent struct {
	// Map of key-value variables used for string replacement in the devfile. Values can be referenced via {{variable-key}}
	// to replace the corresponding value in string fields in the devfile. Replacement cannot be used for
	//
	//  - schemaVersion, metadata, parent source
	//
	//  - element identifiers, e.g. command id, component name, endpoint name, project name
	//
	//  - references to identifiers, e.g. in events, a command's component, container's volume mount name
	//
	//  - string enums, e.g. command group kind, endpoint exposure
	// +optional
	// +patchStrategy=merge
	// +devfile:overrides:include:omitInPlugin=true,description=Overrides of variables encapsulated in a parent devfile.
	Variables map[string]string `json:"variables,omitempty" patchStrategy:"merge"`

	// Map of implementation-dependant free-form YAML attributes.
	// +optional
	// +patchStrategy=merge
	// +devfile:overrides:include:omitInPlugin=true,description=Overrides of attributes encapsulated in a parent devfile.
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Attributes attributes.Attributes `json:"attributes,omitempty" patchStrategy:"merge"`

	// List of the devworkspace components, such as editor and plugins,
	// user-provided containers, or other types of components
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	// +devfile:overrides:include:description=Overrides of components encapsulated in a parent devfile or a plugin.
	// +devfile:toplevellist
	Components []Component `json:"components,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// Projects worked on in the devworkspace, containing names and sources locations
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	// +devfile:overrides:include:omitInPlugin=true,description=Overrides of projects encapsulated in a parent devfile.
	// +devfile:toplevellist
	Projects []Project `json:"projects,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// StarterProjects is a project that can be used as a starting point when bootstrapping new projects
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	// +devfile:overrides:include:omitInPlugin=true,description=Overrides of starterProjects encapsulated in a parent devfile.
	// +devfile:toplevellist
	StarterProjects []StarterProject `json:"starterProjects,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// Predefined, ready-to-use, devworkspace-related commands
	// +optional
	// +patchMergeKey=id
	// +patchStrategy=merge
	// +devfile:overrides:include:description=Overrides of commands encapsulated in a parent devfile or a plugin.
	// +devfile:toplevellist
	Commands []Command `json:"commands,omitempty" patchStrategy:"merge" patchMergeKey:"id"`

	// Bindings of commands to events.
	// Each command is referred-to by its name.
	// +optional
	// +devfile:overrides:include:omit=true
	Events *Events `json:"events,omitempty"`
}
