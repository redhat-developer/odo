package devfile

import (
	attributes "github.com/devfile/api/v2/pkg/attributes"
)

// DevfileHeader describes the structure of the devfile-specific top-level fields
// that are not part of the K8S API structures
type DevfileHeader struct {
	// Devfile schema version
	// +kubebuilder:validation:Pattern=^([2-9])\.([0-9]+)\.([0-9]+)(\-[0-9a-z-]+(\.[0-9a-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$
	SchemaVersion string `json:"schemaVersion"`

	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	// Optional metadata
	Metadata DevfileMetadata `json:"metadata,omitempty"`
}

// Architecture describes the architecture type
// +kubebuilder:validation:Enum=amd64;arm64;ppc64le;s390x
type Architecture string

const (
	AMD64   Architecture = "amd64"
	ARM64   Architecture = "arm64"
	PPC64LE Architecture = "ppc64le"
	S390X   Architecture = "s390x"
)

type DevfileMetadata struct {
	// Optional devfile name
	// +optional
	Name string `json:"name,omitempty"`

	// Optional semver-compatible version
	// +optional
	// +kubebuilder:validation:Pattern=^([0-9]+)\.([0-9]+)\.([0-9]+)(\-[0-9a-z-]+(\.[0-9a-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$
	Version string `json:"version,omitempty"`

	// Map of implementation-dependant free-form YAML attributes. Deprecated, use the top-level attributes field instead.
	// +optional
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Attributes attributes.Attributes `json:"attributes,omitempty"`

	// Optional devfile display name
	// +optional
	DisplayName string `json:"displayName,omitempty"`

	// Optional devfile description
	// +optional
	Description string `json:"description,omitempty"`

	// Optional devfile tags
	// +optional
	Tags []string `json:"tags,omitempty"`

	// Optional list of processor architectures that the devfile supports, empty list suggests that the devfile can be used on any architecture
	// +optional
	// +kubebuilder:validation:UniqueItems=true
	Architectures []Architecture `json:"architectures,omitempty"`

	// Optional devfile icon, can be a URI or a relative path in the project
	// +optional
	Icon string `json:"icon,omitempty"`

	// Optional devfile global memory limit
	// +optional
	GlobalMemoryLimit string `json:"globalMemoryLimit,omitempty"`

	// Optional devfile project type
	// +optional
	ProjectType string `json:"projectType,omitempty"`

	// Optional devfile language
	// +optional
	Language string `json:"language,omitempty"`

	// Optional devfile website
	// +optional
	Website string `json:"website,omitempty"`

	// Optional devfile provider information
	// +optional
	Provider string `json:"provider,omitempty"`

	// Optional link to a page that provides support information
	// +optional
	SupportUrl string `json:"supportUrl,omitempty"`
}
