package v1alpha2

import (
	attributes "github.com/devfile/api/v2/pkg/attributes"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

type Project struct {
	// Project name
	// +kubebuilder:validation:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	// +kubebuilder:validation:MaxLength=63
	Name string `json:"name"`

	// Map of implementation-dependant free-form YAML attributes.
	// +optional
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Attributes attributes.Attributes `json:"attributes,omitempty"`

	// Path relative to the root of the projects to which this project should be cloned into. This is a unix-style relative path (i.e. uses forward slashes). The path is invalid if it is absolute or tries to escape the project root through the usage of '..'. If not specified, defaults to the project name.
	// +optional
	ClonePath string `json:"clonePath,omitempty"`

	ProjectSource `json:",inline"`
}

type StarterProject struct {
	// Project name
	// +kubebuilder:validation:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	// +kubebuilder:validation:MaxLength=63
	Name string `json:"name"`

	// Map of implementation-dependant free-form YAML attributes.
	// +optional
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Attributes attributes.Attributes `json:"attributes,omitempty"`

	// Description of a starter project
	// +optional
	Description string `json:"description,omitempty"`

	// Sub-directory from a starter project to be used as root for starter project.
	// +optional
	SubDir string `json:"subDir,omitempty"`

	ProjectSource `json:",inline"`
}

// ProjectSourceType describes the type of Project sources.
// Only one of the following project sources may be specified.
// If none of the following policies is specified, the default one
// is AllowConcurrent.
// +kubebuilder:validation:Enum=Git;Zip;Custom
type ProjectSourceType string

const (
	GitProjectSourceType    ProjectSourceType = "Git"
	ZipProjectSourceType    ProjectSourceType = "Zip"
	CustomProjectSourceType ProjectSourceType = "Custom"
)

// +union
type ProjectSource struct {
	// Type of project source
	// +
	// +unionDiscriminator
	// +optional
	SourceType ProjectSourceType `json:"sourceType,omitempty"`

	// Project's Git source
	// +optional
	Git *GitProjectSource `json:"git,omitempty"`

	// Project's Zip source
	// +optional
	Zip *ZipProjectSource `json:"zip,omitempty"`

	// Project's Custom source
	// +optional
	// +devfile:overrides:include:omit=true
	Custom *CustomProjectSource `json:"custom,omitempty"`
}

type CommonProjectSource struct {
}

type CustomProjectSource struct {
	ProjectSourceClass string `json:"projectSourceClass"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:EmbeddedResource
	EmbeddedResource runtime.RawExtension `json:"embeddedResource"`
}

type ZipProjectSource struct {
	CommonProjectSource `json:",inline"`

	// Zip project's source location address. Should be file path of the archive, e.g. file://$FILE_PATH
	// +required
	Location string `json:"location,omitempty"`
}

type GitLikeProjectSource struct {
	CommonProjectSource `json:",inline"`

	// Defines from what the project should be checked out. Required if there are more than one remote configured
	// +optional
	CheckoutFrom *CheckoutFrom `json:"checkoutFrom,omitempty"`

	// The remotes map which should be initialized in the git project.
	// Projects must have at least one remote configured while StarterProjects & Image Component's Git source can only have at most one remote configured.
	Remotes map[string]string `json:"remotes"`
}

type CheckoutFrom struct {
	// The revision to checkout from. Should be branch name, tag or commit id.
	// Default branch is used if missing or specified revision is not found.
	// +optional
	Revision string `json:"revision,omitempty"`
	// The remote name should be used as init. Required if there are more than one remote configured
	// +optional
	Remote string `json:"remote,omitempty"`
}

type GitProjectSource struct {
	GitLikeProjectSource `json:",inline"`
}
