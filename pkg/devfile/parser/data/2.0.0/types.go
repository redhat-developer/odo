package version200

import "github.com/openshift/odo/pkg/devfile/parser/data/common"

// CommandGroupType describes the kind of command group.
// +kubebuilder:validation:Enum=build;run;test;debug
type CommandGroupType string

const (
	BuildCommandGroupType CommandGroupType = "build"
	RunCommandGroupType   CommandGroupType = "run"
	TestCommandGroupType  CommandGroupType = "test"
	DebugCommandGroupType CommandGroupType = "debug"
)

// Devfile200 Devfile schema.
type Devfile200 struct {

	// Predefined, ready-to-use, workspace-related commands
	Commands []common.DevfileCommand `json:"commands,omitempty" yaml:"commands,omitempty"`

	// List of the workspace components, such as editor and plugins, user-provided containers, or other types of components
	Components []common.DevfileComponent `json:"components,omitempty" yaml:"components,omitempty"`

	// Bindings of commands to events. Each command is referred-to by its name.
	Events common.DevfileEvents `json:"events,omitempty" yaml:"events,omitempty"`

	// Optional metadata
	Metadata common.DevfileMetadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// Parent workspace template
	Parent common.DevfileParent `json:"parent,omitempty" yaml:"parent,omitempty"`

	// Projects worked on in the workspace, containing names and sources locations
	Projects []common.DevfileProject `json:"projects,omitempty" yaml:"projects,omitempty"`

	// Devfile schema version
	SchemaVersion string `json:"schemaVersion" yaml:"schemaVersion"`

	// StarterProjects is a project that can be used as a starting point when bootstrapping new projects
	StarterProjects []StarterProject `json:"starterProjects, yaml:"schemaVersion", "omitempty"`
}

// Command
type Command struct {

	// Command that consists in applying a given component definition, typically bound to a workspace event.
	//
	// For example, when an `apply` command is bound to a `preStart` event, and references a `container` component, it will start the container as a K8S initContainer in the workspace POD, unless the component has its `dedicatedPod` field set to `true`.
	//
	// When no `apply` command exist for a given component, it is assumed the component will be applied at workspace start by default.
	Apply *Apply `json:"apply,omitempty"`

	// Composite command that allows executing several sub-commands either sequentially or concurrently
	Composite *Composite `json:"composite,omitempty"`

	// CLI Command executed in a component container
	Exec *Exec `json:"exec,omitempty"`

	// Command providing the definition of a VsCode launch action
	VscodeLaunch *VscodeLaunch `json:"vscodeLaunch,omitempty"`

	// Command providing the definition of a VsCode Task
	VscodeTask *VscodeTask `json:"vscodeTask,omitempty"`
}

// Component contains all the configuration related to component
type Component struct {

	// Allows adding and configuring workspace-related containers
	Container *Container `json:"container,omitempty"`

	// Allows importing into the workspace the Kubernetes resources defined in a given manifest. For example this allows reusing the Kubernetes definitions used to deploy some runtime components in production.
	Kubernetes *Kubernetes `json:"kubernetes,omitempty"`

	// Allows importing into the workspace the OpenShift resources defined in a given manifest. For example this allows reusing the OpenShift definitions used to deploy some runtime components in production.
	Openshift *Openshift `json:"openshift,omitempty"`

	// Allows importing a plugin. Plugins are mainly imported devfiles that contribute components, commands and events as a consistent single unit. They are defined in either YAML files following the devfile syntax, or as `DevWorkspaceTemplate` Kubernetes Custom Resources
	Plugin *Plugin `json:"plugin,omitempty"`

	// Allows specifying the definition of a volume shared by several other components
	Volume *Volume `json:"volume,omitempty"`
}

// Composite Composite command that allows executing several sub-commands either sequentially or concurrently
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
	Label string `json:"label,omitempty"`

	// Indicates if the sub-commands should be executed concurrently
	Parallel bool `json:"parallel,omitempty"`
}

// Configuration holds configuration for an endpoint
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

// Container Allows adding and configuring workspace-related containers
type Container struct {

	// The arguments to supply to the command running the dockerimage component. The arguments are supplied either to the default command provided in the image or to the overridden command. Defaults to an empty array, meaning use whatever is defined in the image.
	Args []string `json:"args,omitempty"`

	// The command to run in the dockerimage component instead of the default one provided in the image. Defaults to an empty array, meaning use whatever is defined in the image.
	Command []string `json:"command,omitempty"`

	Endpoints []*Endpoint `json:"endpoints,omitempty"`

	// Specify if a container should run in its own separated pod, instead of running as part of the main development environment pod.
	//
	// Default value is `false`
	DedicatedPod bool `json:"dedicatedPod,omitempty"`

	// Environment variables used in this container
	Env          []*Env `json:"env,omitempty"`
	Image        string `json:"image,omitempty"`
	MemoryLimit  string `json:"memoryLimit,omitempty"`
	MountSources bool   `json:"mountSources,omitempty"`
	Name         string `json:"name"`

	// Optional specification of the path in the container where project sources should be transferred/mounted when `mountSources` is `true`. When omitted, the value of the `PROJECTS_ROOT` environment variable is used.
	SourceMapping string `json:"sourceMapping,omitempty"`

	// List of volumes mounts that should be mounted is this container.
	VolumeMounts []*VolumeMount `json:"volumeMounts,omitempty"`
}

// Endpoint holds information about how an application is exposed
type Endpoint struct {

	// Map of implementation-dependant string-based free-form attributes.
	//
	// Examples of Che-specific attributes:
	// - cookiesAuthEnabled: "true" / "false",
	// - type: "terminal" / "ide" / "ide-dev",
	Attributes map[string]string `json:"attributes,omitempty"`

	// Describes how the endpoint should be exposed on the network.
	// - `public` means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.
	// - `internal` means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.
	// - `none` means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.
	//
	// Default value is `public`
	Exposure string `json:"exposure,omitempty"`
	Name     string `json:"name"`

	// Path of the endpoint URL
	Path string `json:"path,omitempty"`

	// Describes the application and transport protocols of the traffic that will go through this endpoint.
	// - `http`: Endpoint will have `http` traffic, typically on a TCP connection. It will be automaticaly promoted to `https` when the `secure` field is set to `true`.
	// - `https`: Endpoint will have `https` traffic, typically on a TCP connection.
	// - `ws`: Endpoint will have `ws` traffic, typically on a TCP connection. It will be automaticaly promoted to `wss` when the `secure` field is set to `true`.
	// - `wss`: Endpoint will have `wss` traffic, typically on a TCP connection.
	// - `tcp`: Endpoint will have traffic on a TCP connection, without specifying an application protocol.
	// - `udp`: Endpoint will have traffic on an UDP connection, without specifying an application protocol.
	//
	// Default value is `http`
	Protocol string `json:"protocol,omitempty"`

	// Describes whether the endpoint should be secured and protected by some authentication process
	Secure     bool  `json:"secure,omitempty"`
	TargetPort int32 `json:"targetPort,omitempty"`
}

// Env is the key value pair representing an Environment variable
type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Events Bindings of commands to events. Each command is referred-to by its name.
type Events struct {

	// Names of commands that should be executed after the workspace is completely started. In the case of Che-Theia, these commands should be executed after all plugins and extensions have started, including project cloning. This means that those commands are not triggered until the user opens the IDE in his browser.
	PostStart []string `json:"postStart,omitempty"`

	// Names of commands that should be executed after stopping the workspace.
	PostStop []string `json:"postStop,omitempty"`

	// Names of commands that should be executed before the workspace start. Kubernetes-wise, these commands would typically be executed in init containers of the workspace POD.
	PreStart []string `json:"preStart,omitempty"`

	// Names of commands that should be executed before stopping the workspace.
	PreStop []string `json:"preStop,omitempty"`
}

// Exec CLI Command executed in a component container
type Exec struct {

	// Optional map of free-form additional command attributes
	Attributes map[string]string `json:"attributes,omitempty"`

	// The actual command-line string
	CommandLine string `json:"commandLine,omitempty"`

	// Describes component to which given action relates
	Component string `json:"component,omitempty"`

	// Optional list of environment variables that have to be set before running the command
	Env []*Env `json:"env,omitempty"`

	// Defines the group this command is part of
	Group *Group `json:"group,omitempty"`

	// Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.
	Id string `json:"id"`

	// Optional label that provides a label for this command to be used in Editor UI menus for example
	Label string `json:"label,omitempty"`

	// Working directory where the command should be executed
	WorkingDir string `json:"workingDir,omitempty"`
}

// Apply Command that consists in applying a given component definition, typically bound to a workspace event.
//
// For example, when an `apply` command is bound to a `preStart` event, and references a `container` component, it will start the container as a K8S initContainer in the workspace POD, unless the component has its `dedicatedPod` field set to `true`.
//
// When no `apply` command exist for a given component, it is assumed the component will be applied at workspace start by default.
type Apply struct {

	// Optional map of free-form additional command attributes
	Attributes map[string]string `json:"attributes,omitempty"`

	// Describes component that will be applied
	Component string `json:"component,omitempty"`

	// Defines the group this command is part of
	Group *Group `json:"group,omitempty"`

	// Mandatory identifier that allows referencing this command in composite commands, from a parent, or in events.
	Id string `json:"id"`

	// Optional label that provides a label for this command to be used in Editor UI menus for example
	Label string `json:"label,omitempty"`
}

// Git Project's Git source
type Git struct {

	// The branch to check
	Branch string `json:"branch,omitempty"`

	// Project's source location address. Should be URL for git and github located projects, or; file:// for zip
	Location string `json:"location,omitempty"`

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
	Location string `json:"location,omitempty"`

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
	Kind CommandGroupType `json:"kind"`
}

// Kubernetes Allows importing into the workspace the Kubernetes resources defined in a given manifest. For example this allows reusing the Kubernetes definitions used to deploy some runtime components in production.
type Kubernetes struct {

	// Inlined manifest
	Inlined string `json:"inlined,omitempty"`

	// Mandatory name that allows referencing the component in commands, or inside a parent
	Name string `json:"name"`

	// Location in a file fetched from a uri.
	Uri string `json:"uri,omitempty"`
}

// Metadata Optional metadata
type Metadata struct {

	// Optional devfile name
	Name string `json:"name,omitempty"`

	// Optional semver-compatible version
	Version string `json:"version,omitempty"`
}

// Openshift Configuration overriding for an OpenShift component
type Openshift struct {

	// Inlined manifest
	Inlined string `json:"inlined,omitempty"`

	// Mandatory name that allows referencing the component in commands, or inside a parent
	Name string `json:"name"`

	// Location in a file fetched from a uri.
	Uri string `json:"uri,omitempty"`
}

// Parent Parent workspace template
type Parent struct {

	// Predefined, ready-to-use, workspace-related commands
	Commands []Command `json:"commands,omitempty"`

	// List of the workspace components, such as editor and plugins, user-provided containers, or other types of components
	Components []Component `json:"components,omitempty"`

	// Bindings of commands to events. Each command is referred-to by its name.
	Events Events `json:"events,omitempty"`

	// Id in a registry that contains a Devfile yaml file
	Id string `json:"id,omitempty"`

	// Reference to a Kubernetes CRD of type DevWorkspaceTemplate
	Kubernetes Kubernetes `json:"kubernetes,omitempty"`

	// Projects worked on in the workspace, containing names and sources locations
	Projects    []Project `json:"projects,omitempty"`
	RegistryUrl string    `json:"registryUrl,omitempty"`

	// StarterProjects is a project that can be used as a starting point when bootstrapping new projects
	StarterProjects []StarterProject `json:"starterProjects,omitempty"`

	// Uri of a Devfile yaml file
	Uri string `json:"uri,omitempty"`
}

// Plugin Allows importing a plugin. Plugins are mainly imported devfiles that contribute components, commands and events as a consistent single unit. They are defined in either YAML files following the devfile syntax, or as `DevWorkspaceTemplate` Kubernetes Custom Resources
type Plugin struct {

	// Overrides of commands encapsulated in a plugin. Overriding is done using a strategic merge
	Commands []*Command `json:"commands,omitempty"`

	// Overrides of components encapsulated in a plugin. Overriding is done using a strategic merge
	Components []*Component `json:"components,omitempty"`

	// Id in a registry that contains a Devfile yaml file
	Id string `json:"id,omitempty"`

	// Reference to a Kubernetes CRD of type DevWorkspaceTemplate
	Kubernetes *Kubernetes `json:"kubernetes,omitempty"`

	// Optional name that allows referencing the component in commands, or inside a parent If omitted it will be infered from the location (uri or registryEntry)
	Name        string `json:"name,omitempty"`
	RegistryUrl string `json:"registryUrl,omitempty"`

	// Uri of a Devfile yaml file
	Uri string `json:"uri,omitempty"`
}

// Project holds details of a starter project that can be downloaded by the user
type Project struct {

	// Path relative to the root of the projects to which this project should be cloned into. This is a unix-style relative path (i.e. uses forward slashes). The path is invalid if it is absolute or tries to escape the project root through the usage of '..'. If not specified, defaults to the project name.
	ClonePath string `json:"clonePath,omitempty"`

	// Project's Git source
	Git *Git `json:"git,omitempty"`

	// Project's GitHub source
	Github *Github `json:"github,omitempty"`

	// Project name
	Name string `json:"name"`

	// Project's Zip source
	Zip *Zip `json:"zip,omitempty"`
}

// StarterProject
type StarterProject struct {

	// Path relative to the root of the projects to which this project should be cloned into. This is a unix-style relative path (i.e. uses forward slashes). The path is invalid if it is absolute or tries to escape the project root through the usage of '..'. If not specified, defaults to the project name.
	ClonePath string `json:"clonePath,omitempty"`

	// Description of a starter project
	Description string `json:"description,omitempty"`

	// Project's Git source
	Git *Git `json:"git,omitempty"`

	// Project's GitHub source
	Github *Github `json:"github,omitempty"`

	// Description of a starter project
	MarkdownDescription string `json:"markdownDescription,omitempty"`

	// Project name
	Name string `json:"name"`

	// Project's Zip source
	Zip *Zip `json:"zip,omitempty"`
}

// Volume Allows specifying the definition of a volume shared by several other components
type Volume struct {

	// Mandatory name that allows referencing the Volume component in Container volume mounts or inside a parent
	Name string `json:"name"`

	// Size of the volume
	Size string `json:"size,omitempty"`
}

// VolumeMountsItems Volume that should be mounted to a component container
type VolumeMount struct {

	// The volume mount name is the name of an existing `Volume` component. If several containers mount the same volume name then they will reuse the same volume and will be able to access to the same files.
	Name string `json:"name"`

	// The path in the component container where the volume should be mounted. If not path is mentioned, default path is the is `/<name>`.
	Path string `json:"path,omitempty"`
}

// VscodeLaunch Command providing the definition of a VsCode launch action
type VscodeLaunch struct {

	// Optional map of free-form additional command attributes
	Attributes map[string]string `json:"attributes,omitempty"`

	// Defines the group this command is part of
	Group *Group `json:"group,omitempty"`

	// Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.
	Id string `json:"id"`

	// Inlined content of the VsCode configuration
	Inlined string `json:"inlined,omitempty"`

	// Location as an absolute of relative URI the VsCode configuration will be fetched from
	Uri string `json:"uri,omitempty"`
}

// VscodeTask Command providing the definition of a VsCode Task
type VscodeTask struct {

	// Optional map of free-form additional command attributes
	Attributes map[string]string `json:"attributes,omitempty"`

	// Defines the group this command is part of
	Group *Group `json:"group,omitempty"`

	// Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.
	Id string `json:"id"`

	// Inlined content of the VsCode configuration
	Inlined string `json:"inlined,omitempty"`

	// Location as an absolute of relative URI the VsCode configuration will be fetched from
	Uri string `json:"uri,omitempty"`
}

// Zip Project's Zip source
type Zip struct {

	// Project's source location address. Should be URL for git and github located projects, or; file:// for zip
	Location string `json:"location,omitempty"`

	// Part of project to populate in the working directory.
	SparseCheckoutDir string `json:"sparseCheckoutDir,omitempty"`
}
