package common

// ProjectSourceType describes the type of Project sources.
// Only one of the following project sources may be specified.
type DevfileProjectSourceType string

const (
	GitProjectSourceType    DevfileProjectSourceType = "Git"
	GitHubProjectSourceType DevfileProjectSourceType = "Github"
	ZipProjectSourceType    DevfileProjectSourceType = "Zip"
	CustomProjectSourceType DevfileProjectSourceType = "Custom"
)

type DevfileComponentType string

const (
	ContainerComponentType  DevfileComponentType = "Container"
	KubernetesComponentType DevfileComponentType = "Kubernetes"
	OpenshiftComponentType  DevfileComponentType = "Openshift"
	PluginComponentType     DevfileComponentType = "Plugin"
	VolumeComponentType     DevfileComponentType = "Volume"
	CustomComponentType     DevfileComponentType = "Custom"
)

type DevfileCommandType string

const (
	ExecCommandType         DevfileCommandType = "Exec"
	VscodeTaskCommandType   DevfileCommandType = "VscodeTask"
	VscodeLaunchCommandType DevfileCommandType = "VscodeLaunch"
	CompositeCommandType    DevfileCommandType = "Composite"
	CustomCommandType       DevfileCommandType = "Custom"
)

// CommandGroupType describes the kind of command group.
// +kubebuilder:validation:Enum=build;run;test;debug
type DevfileCommandGroupType string

const (
	BuildCommandGroupType DevfileCommandGroupType = "build"
	RunCommandGroupType   DevfileCommandGroupType = "run"
	TestCommandGroupType  DevfileCommandGroupType = "test"
	DebugCommandGroupType DevfileCommandGroupType = "debug"
	// To Support V1
	InitCommandGroupType DevfileCommandGroupType = "init"
)

// Metadata Optional metadata
type DevfileMetadata struct {

	// Optional devfile name
	Name string `json:"name,omitempty"`

	// Optional semver-compatible version
	Version string `json:"version,omitempty"`
}

// CommandsItems
type DevfileCommand struct {

	// Exec command
	Exec *Exec `json:"exec,omitempty"`

	// Type of workspace command
	Type DevfileCommandType `json:"type,omitempty"`
}

// ComponentsItems
type DevfileComponent struct {

	// CheEditor component
	CheEditor *CheEditor `json:"cheEditor,omitempty"`

	// ChePlugin component
	ChePlugin *ChePlugin `json:"chePlugin,omitempty"`

	// Container component
	Container *Container `json:"container,omitempty"`

	// Custom component
	Custom *Custom `json:"custom,omitempty"`

	// Kubernetes component
	Kubernetes *Kubernetes `json:"kubernetes,omitempty"`

	// Openshift component
	Openshift *Openshift `json:"openshift,omitempty"`

	// Type of project source
	Type DevfileComponentType `json:"type,omitempty"`

	// Volume component
	Volume *Volume `json:"volume,omitempty"`
}

// ProjectsItems
type DevfileProject struct {

	// Path relative to the root of the projects to which this project should be cloned into. This is a unix-style relative path (i.e. uses forward slashes). The path is invalid if it is absolute or tries to escape the project root through the usage of '..'. If not specified, defaults to the project name.
	ClonePath string `json:"clonePath,omitempty"`

	// Project's Custom source
	Custom *Custom `json:"custom,omitempty"`

	// Project's Git source
	Git *Git `json:"git,omitempty"`

	// Project's GitHub source
	Github *Github `json:"github,omitempty"`

	// Project name
	Name string `json:"name"`

	// Type of project source
	SourceType DevfileProjectSourceType `json:"sourceType,omitempty"`

	// Project's Zip source
	Zip *Zip `json:"zip,omitempty"`
}

// CheEditor CheEditor component
type CheEditor struct {

	// Type of plugin location
	LocationType string `json:"locationType,omitempty"`
	MemoryLimit  string `json:"memoryLimit,omitempty"`

	// Optional name that allows referencing the component in commands, or inside a parent If omitted it will be infered from the location (uri or registryEntry)
	Name string `json:"name,omitempty"`

	// Location of an entry inside a plugin registry
	RegistryEntry *RegistryEntry `json:"registryEntry,omitempty"`

	// Location defined as an URI
	Uri string `json:"uri,omitempty"`
}

// ChePlugin ChePlugin component
type ChePlugin struct {

	// Type of plugin location
	LocationType string `json:"locationType,omitempty"`
	MemoryLimit  string `json:"memoryLimit,omitempty"`

	// Optional name that allows referencing the component in commands, or inside a parent If omitted it will be infered from the location (uri or registryEntry)
	Name string `json:"name,omitempty"`

	// Location of an entry inside a plugin registry
	RegistryEntry *RegistryEntry `json:"registryEntry,omitempty"`

	// Location defined as an URI
	Uri string `json:"uri,omitempty"`
}

// Composite Composite command
type Composite struct {

	// Optional map of free-form additional command attributes
	Attributes map[string]string `json:"attributes,omitempty"`

	// The commands that comprise this composite command
	Commands []string `json:"commands,omitempty"`

	// Defines the group this command is part of
	Group *Group `json:"group,omitempty"`

	// Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.
	Id string `json:"id"`

	// Optional label that provides a label for this command to be used in Editor UI menus for example
	Label    string `json:"label,omitempty"`
	Parallel bool   `json:"parallel,omitempty"`
}

// Configuration
type Configuration struct {
	CookiesAuthEnabled bool   `json:"cookiesAuthEnabled,omitempty"`
	Discoverable       bool   `json:"discoverable,omitempty"`
	Path               string `json:"path,omitempty"`

	// The is the low-level protocol of traffic coming through this endpoint. Default value is "tcp"
	Protocol string `json:"protocol,omitempty"`
	Public   bool   `json:"public,omitempty"`

	// The is the URL scheme to use when accessing the endpoint. Default value is "http"
	Scheme string `json:"scheme,omitempty"`
	Secure bool   `json:"secure,omitempty"`
	Type   string `json:"type,omitempty"`
}

// Container Container component
type Container struct {
	Endpoints []Endpoint `json:"endpoints,omitempty"`

	// Environment variables used in this container
	Env          []Env  `json:"env,omitempty"`
	Image        string `json:"image"`
	MemoryLimit  string `json:"memoryLimit,omitempty"`
	MountSources bool   `json:"mountSources,omitempty"`
	Name         string `json:"name"`

	// Optional specification of the path in the container where project sources should be transferred/mounted when `mountSources` is `true`. When omitted, the value of the `PROJECTS_ROOT` environment variable is used.
	SourceMapping string `json:"sourceMapping,omitempty"`

	// List of volumes mounts that should be mounted is this container.
	VolumeMounts []VolumeMount `json:"volumeMounts,omitempty"`

	Command []string `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
}

// Custom Custom component
type Custom struct {
	ComponentClass   string            `json:"componentClass"`
	EmbeddedResource *EmbeddedResource `json:"embeddedResource"`
	Name             string            `json:"name"`
}

// EmbeddedResource
type EmbeddedResource struct {
}

// Endpoint
type Endpoint struct {
	Attributes    map[string]string `json:"attributes,omitempty"`
	Configuration *Configuration    `json:"configuration"`
	Name          string            `json:"name"`
	TargetPort    int32             `json:"targetPort"`
}

// Env
type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Events Bindings of commands to events. Each command is referred-to by its name.
type DevfileEvents struct {

	// Names of commands that should be executed after the workspace is completely started. In the case of Che-Theia, these commands should be executed after all plugins and extensions have started, including project cloning. This means that those commands are not triggered until the user opens the IDE in his browser.
	PostStart []string `json:"postStart,omitempty"`

	// Names of commands that should be executed after stopping the workspace.
	PostStop []string `json:"postStop,omitempty"`

	// Names of commands that should be executed before the workspace start. Kubernetes-wise, these commands would typically be executed in init containers of the workspace POD.
	PreStart []string `json:"preStart,omitempty"`

	// Names of commands that should be executed before stopping the workspace.
	PreStop []string `json:"preStop,omitempty"`
}

// Exec Exec command
type Exec struct {

	// Optional map of free-form additional command attributes
	Attributes map[string]string `json:"attributes,omitempty"`

	// The actual command-line string
	CommandLine string `json:"commandLine"`

	// Describes component to which given action relates
	Component string `json:"component,omitempty"`

	// Optional list of environment variables that have to be set before running the command
	Env []Env `json:"env,omitempty"`

	// Defines the group this command is part of
	Group *Group `json:"group,omitempty"`

	// Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.
	Id string `json:"id"`

	// Optional label that provides a label for this command to be used in Editor UI menus for example
	Label string `json:"label,omitempty"`

	// Working directory where the command should be executed
	WorkingDir string `json:"workingDir,omitempty"`
}

// Git Project's Git source
type Git struct {

	// The branch to check
	Branch string `json:"branch,omitempty"`

	// Project's source location address. Should be URL for git and github located projects, or; file:// for zip
	Location string `json:"location"`

	// Part of project to populate in the working directory.
	SparseCheckoutDir string `json:"sparseCheckoutDir,omitempty"`

	// The tag or commit id to reset the checked out branch to
	StartPoint string `json:"startPoint,omitempty"`
}

// Github Project's GitHub source
type Github struct {

	// The branch to check
	Branch string `json:"branch,omitempty"`

	// Project's source location address. Should be URL for git and github located projects, or; file:// for zip
	Location string `json:"location"`

	// Part of project to populate in the working directory.
	SparseCheckoutDir string `json:"sparseCheckoutDir,omitempty"`

	// The tag or commit id to reset the checked out branch to
	StartPoint string `json:"startPoint,omitempty"`
}

// Group Defines the group this command is part of
type Group struct {

	// Identifies the default command for a given group kind
	IsDefault bool `json:"isDefault,omitempty"`

	// Kind of group the command is part of
	Kind DevfileCommandGroupType `json:"kind"`
}

// Kubernetes Kubernetes component
type Kubernetes struct {

	// Reference to the plugin definition
	Inlined string `json:"inlined,omitempty"`

	// Type of Kubernetes-like location
	LocationType string `json:"locationType,omitempty"`

	// Mandatory name that allows referencing the component in commands, or inside a parent
	Name string `json:"name"`

	// Location in a plugin registry
	Url string `json:"url,omitempty"`
}

// Openshift Openshift component
type Openshift struct {

	// Reference to the plugin definition
	Inlined string `json:"inlined,omitempty"`

	// Type of Kubernetes-like location
	LocationType string `json:"locationType,omitempty"`

	// Mandatory name that allows referencing the component in commands, or inside a parent
	Name string `json:"name"`

	// Location in a plugin registry
	Url string `json:"url,omitempty"`
}

// Parent Parent workspace template
type DevfileParent struct {

	// Reference to a Kubernetes CRD of type DevWorkspaceTemplate
	Kubernetes *Kubernetes `json:"kubernetes,omitempty"`

	// Type of parent location
	LocationType string `json:"locationType,omitempty"`

	// Entry in a registry (base URL + ID) that contains a Devfile yaml file
	RegistryEntry *RegistryEntry `json:"registryEntry,omitempty"`

	// Uri of a Devfile yaml file
	Uri string `json:"uri,omitempty"`
}

// RegistryEntry Location of an entry inside a plugin registry
type RegistryEntry struct {
	BaseUrl string `json:"baseUrl,omitempty"`
	Id      string `json:"id"`
}

// Volume Volume component
type Volume struct {

	// Mandatory name that allows referencing the Volume component in Container volume mounts or inside a parent
	Name string `json:"name"`

	// Size of the volume
	Size string `json:"size,omitempty"`
}

// VolumeMountsItems Volume that should be mounted to a component container
type VolumeMount struct {

	// The volume mount name is the name of an existing `Volume` component. If no corresponding `Volume` component exist it is implicitly added. If several containers mount the same volume name then they will reuse the same volume and will be able to access to the same files.
	Name string `json:"name"`

	// The path in the component container where the volume should be mounted
	Path string `json:"path"`
}

// VscodeLaunch VscodeLaunch command
type VscodeLaunch struct {

	// Optional map of free-form additional command attributes
	Attributes map[string]string `json:"attributes,omitempty"`

	// Defines the group this command is part of
	Group *Group `json:"group,omitempty"`

	// Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.
	Id string `json:"id"`

	// Embedded content of the vscode configuration file
	Inlined string `json:"inlined,omitempty"`

	// Type of Vscode configuration command location
	LocationType string `json:"locationType,omitempty"`

	// Location as an absolute of relative URL
	Url string `json:"url,omitempty"`
}

// VscodeTask VscodeTask command
type VscodeTask struct {

	// Optional map of free-form additional command attributes
	Attributes map[string]string `json:"attributes,omitempty"`

	// Defines the group this command is part of
	Group *Group `json:"group,omitempty"`

	// Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.
	Id string `json:"id"`

	// Embedded content of the vscode configuration file
	Inlined string `json:"inlined,omitempty"`

	// Type of Vscode configuration command location
	LocationType string `json:"locationType,omitempty"`

	// Location as an absolute of relative URL
	Url string `json:"url,omitempty"`
}

// Zip Project's Zip source
type Zip struct {

	// Project's source location address. Should be URL for git and github located projects, or; file:// for zip
	Location string `json:"location"`

	// Part of project to populate in the working directory.
	SparseCheckoutDir string `json:"sparseCheckoutDir,omitempty"`
}
