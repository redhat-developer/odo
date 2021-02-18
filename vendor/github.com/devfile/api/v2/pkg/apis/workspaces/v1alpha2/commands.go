package v1alpha2

import (
	attributes "github.com/devfile/api/v2/pkg/attributes"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// CommandType describes the type of command.
// Only one of the following command type may be specified.
// +kubebuilder:validation:Enum=Exec;Apply;VscodeTask;VscodeLaunch;Composite;Custom
type CommandType string

const (
	ExecCommandType         CommandType = "Exec"
	ApplyCommandType        CommandType = "Apply"
	VscodeTaskCommandType   CommandType = "VscodeTask"
	VscodeLaunchCommandType CommandType = "VscodeLaunch"
	CompositeCommandType    CommandType = "Composite"
	CustomCommandType       CommandType = "Custom"
)

// CommandGroupKind describes the kind of command group.
// +kubebuilder:validation:Enum=build;run;test;debug
type CommandGroupKind string

const (
	BuildCommandGroupKind CommandGroupKind = "build"
	RunCommandGroupKind   CommandGroupKind = "run"
	TestCommandGroupKind  CommandGroupKind = "test"
	DebugCommandGroupKind CommandGroupKind = "debug"
)

type CommandGroup struct {
	// Kind of group the command is part of
	Kind CommandGroupKind `json:"kind"`

	// +optional
	// Identifies the default command for a given group kind
	IsDefault bool `json:"isDefault,omitempty"`
}

type BaseCommand struct {
	// +optional
	// Defines the group this command is part of
	Group *CommandGroup `json:"group,omitempty"`
}

type LabeledCommand struct {
	BaseCommand `json:",inline"`

	// +optional
	// Optional label that provides a label for this command
	// to be used in Editor UI menus for example
	Label string `json:"label,omitempty"`
}

type Command struct {
	// Mandatory identifier that allows referencing
	// this command in composite commands, from
	// a parent, or in events.
	// +kubebuilder:validation:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	// +kubebuilder:validation:MaxLength=63
	Id string `json:"id"`
	// Map of implementation-dependant free-form YAML attributes.
	// +optional
	Attributes   attributes.Attributes `json:"attributes,omitempty"`
	CommandUnion `json:",inline"`
}

// +union
type CommandUnion struct {
	// Type of workspace command
	// +unionDiscriminator
	// +optional
	CommandType CommandType `json:"commandType,omitempty"`

	// CLI Command executed in an existing component container
	// +optional
	Exec *ExecCommand `json:"exec,omitempty"`

	// Command that consists in applying a given component definition,
	// typically bound to a workspace event.
	//
	// For example, when an `apply` command is bound to a `preStart` event,
	// and references a `container` component, it will start the container as a
	// K8S initContainer in the workspace POD, unless the component has its
	// `dedicatedPod` field set to `true`.
	//
	// When no `apply` command exist for a given component,
	// it is assumed the component will be applied at workspace start
	// by default.
	// +optional
	Apply *ApplyCommand `json:"apply,omitempty"`

	// Command providing the definition of a VsCode Task
	// +optional
	VscodeTask *VscodeConfigurationCommand `json:"vscodeTask,omitempty"`

	// Command providing the definition of a VsCode launch action
	// +optional
	VscodeLaunch *VscodeConfigurationCommand `json:"vscodeLaunch,omitempty"`

	// Composite command that allows executing several sub-commands
	// either sequentially or concurrently
	// +optional
	Composite *CompositeCommand `json:"composite,omitempty"`

	// Custom command whose logic is implementation-dependant
	// and should be provided by the user
	// possibly through some dedicated plugin
	// +optional
	// +devfile:overrides:include:omit=true
	Custom *CustomCommand `json:"custom,omitempty"`
}

type ExecCommand struct {
	LabeledCommand `json:",inline"`

	// The actual command-line string
	//
	// Special variables that can be used:
	//
	//  - `$PROJECTS_ROOT`: A path where projects sources are mounted as defined by container component's sourceMapping.
	//
	//  - `$PROJECT_SOURCE`: A path to a project source ($PROJECTS_ROOT/<project-name>). If there are multiple projects, this will point to the directory of the first one.
	CommandLine string `json:"commandLine"`

	// Describes component to which given action relates
	//
	Component string `json:"component"`

	// Working directory where the command should be executed
	//
	// Special variables that can be used:
	//
	//  - `$PROJECTS_ROOT`: A path where projects sources are mounted as defined by container component's sourceMapping.
	//
	//  - `$PROJECT_SOURCE`: A path to a project source ($PROJECTS_ROOT/<project-name>). If there are multiple projects, this will point to the directory of the first one.
	// +optional
	WorkingDir string `json:"workingDir,omitempty"`

	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	// Optional list of environment variables that have to be set
	// before running the command
	Env []EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// +optional
	// Whether the command is capable to reload itself when source code changes.
	// If set to `true` the command won't be restarted and it is expected to handle file changes on its own.
	//
	// Default value is `false`
	HotReloadCapable bool `json:"hotReloadCapable,omitempty"`
}

type ApplyCommand struct {
	LabeledCommand `json:",inline"`

	// Describes component that will be applied
	//
	Component string `json:"component"`
}

type CompositeCommand struct {
	LabeledCommand `json:",inline"`

	// The commands that comprise this composite command
	Commands []string `json:"commands,omitempty" patchStrategy:"replace"`

	// Indicates if the sub-commands should be executed concurrently
	// +optional
	Parallel bool `json:"parallel,omitempty"`
}

// VscodeConfigurationCommandLocationType describes the type of
// the location the configuration is fetched from.
// Only one of the following component type may be specified.
// +kubebuilder:validation:Enum=Uri;Inlined
type VscodeConfigurationCommandLocationType string

const (
	UriVscodeConfigurationCommandLocationType     VscodeConfigurationCommandLocationType = "Uri"
	InlinedVscodeConfigurationCommandLocationType VscodeConfigurationCommandLocationType = "Inlined"
)

// +union
type VscodeConfigurationCommandLocation struct {
	// Type of Vscode configuration command location
	// +
	// +unionDiscriminator
	// +optional
	LocationType VscodeConfigurationCommandLocationType `json:"locationType,omitempty"`

	// Location as an absolute of relative URI
	// the VsCode configuration will be fetched from
	// +optional
	Uri string `json:"uri,omitempty"`

	// Inlined content of the VsCode configuration
	// +optional
	Inlined string `json:"inlined,omitempty"`
}

type VscodeConfigurationCommand struct {
	BaseCommand                        `json:",inline"`
	VscodeConfigurationCommandLocation `json:",inline"`
}

type CustomCommand struct {
	LabeledCommand `json:",inline"`

	// Class of command that the associated implementation component
	// should use to process this command with the appropriate logic
	CommandClass string `json:"commandClass"`

	// Additional free-form configuration for this custom command
	// that the implementation component will know how to use
	//
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:EmbeddedResource
	EmbeddedResource runtime.RawExtension `json:"embeddedResource"`
}
