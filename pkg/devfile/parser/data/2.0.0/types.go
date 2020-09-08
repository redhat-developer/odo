package version200

import "github.com/openshift/odo/pkg/devfile/parser/data/common"

// Devfile200 Devfile schema.
type Devfile200 struct {
	// Devfile schema version
	SchemaVersion string `json:"schemaVersion" yaml:"schemaVersion"`

	// Optional metadata
	Metadata common.DevfileMetadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// Parent workspace template
	Parent common.DevfileParent `json:"parent,omitempty" yaml:"parent,omitempty"`

	// Projects worked on in the workspace, containing names and sources locations
	Projects []common.DevfileProject `json:"projects,omitempty" yaml:"projects,omitempty"`

	// StarterProjects is a project that can be used as a starting point when bootstrapping new projects
	StarterProjects []common.DevfileStarterProject `json:"starterProjects,omitempty" yaml:"starterProjects,omitempty"`

	// List of the workspace components, such as editor and plugins, user-provided containers, or other types of components
	Components []common.DevfileComponent `json:"components,omitempty" yaml:"components,omitempty"`

	// Predefined, ready-to-use, workspace-related commands
	Commands []common.DevfileCommand `json:"commands,omitempty" yaml:"commands,omitempty"`

	// Bindings of commands to events. Each command is referred-to by its name.
	Events common.DevfileEvents `json:"events,omitempty" yaml:"events,omitempty"`
}
