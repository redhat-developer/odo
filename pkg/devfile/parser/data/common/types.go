package common

// DevfileComponentType describes the type of component.
// Only one of the following component type may be specified
// To support some print actions
type DevfileComponentType string

const (
	ContainerComponentType  DevfileComponentType = "Container"
	KubernetesComponentType DevfileComponentType = "Kubernetes"
	OpenshiftComponentType  DevfileComponentType = "Openshift"
	PluginComponentType     DevfileComponentType = "Plugin"
	VolumeComponentType     DevfileComponentType = "Volume"
	CustomComponentType     DevfileComponentType = "Custom"
)

// DevfileCommandGroupType describes the kind of command group.
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

// ExposureType is an enum to indicate the exposure type of the endpoint
type ExposureType string

const (
	Public   ExposureType = "public"
	Internal ExposureType = "internal"
	None     ExposureType = "none"
)

// ProtocolType is an enum to indicate the protocol type of the endpoint
type ProtocolType string

const (
	HTTP  ProtocolType = "http"
	HTTPS ProtocolType = "https"
	WS    ProtocolType = "ws"
	WSS   ProtocolType = "wss"
	TCP   ProtocolType = "tcp"
	UDP   ProtocolType = "udp"
)

// DevfileMetadata metadata for devfile
type DevfileMetadata struct {

	// Name Optional devfile name
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Version Optional semver-compatible version
	Version string `json:"version,omitempty"`

	// Dockerfile optional URL to remote Dockerfile
	Dockerfile string `json:"alpha.build-dockerfile,omitempty"`

	// Manifest optional URL to remote Deployment Manifest
	Manifest string `json:"alpha.deployment-manifest,omitempty"`
}

// DevfileCommand command specified in devfile
type DevfileCommand struct {

	// Composite command executed in a component container
	Composite *Composite `json:"composite,omitempty" yaml:"composite,omitempty"`
	// CLI Command executed in a component container
	Exec *Exec `json:"exec,omitempty" yaml:"exec,omitempty"`

	// Mandatory identifier that allows referencing this command in composite commands, from a parent, or in events.
	Id string `json:"id" yaml:"id"`
}

// DevfileComponent component specified in devfile
type DevfileComponent struct {

	// Mandatory name that allows referencing the component from other elements (such as commands) or from an external devfile that may reference this component through a parent or a plugin.
	Name string `json:"name" yaml:"name"`

	// Allows adding and configuring workspace-related containers
	Container *Container `json:"container,omitempty" yaml:"container,omitempty"`

	// Allows importing into the workspace the Kubernetes resources defined in a given manifest. For example this allows reusing the Kubernetes definitions used to deploy some runtime components in production.
	Kubernetes *Kubernetes `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty"`

	// Allows importing into the workspace the OpenShift resources defined in a given manifest. For example this allows reusing the OpenShift definitions used to deploy some runtime components in production.
	Openshift *Openshift `json:"openshift,omitempty" yaml:"openshift,omitempty"`

	// Allows specifying the definition of a volume shared by several other components
	Volume *Volume `json:"volume,omitempty" yaml:"volume,omitempty"`
}

// Container Allows adding and configuring workspace-related containers
type Container struct {

	// The arguments to supply to the command running the dockerimage component. The arguments are supplied either to the default command provided in the image or to the overridden command. Defaults to an empty array, meaning use whatever is defined in the image.
	Args []string `json:"args,omitempty" yaml:"args,omitempty" `

	// The command to run in the dockerimage component instead of the default one provided in the image. Defaults to an empty array, meaning use whatever is defined in the image.
	Command []string `json:"command,omitempty" yaml:"command,omitempty"`

	Endpoints []Endpoint `json:"endpoints,omitempty" yaml:"endpoints,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// Environment variables used in this container
	Env []Env `json:"env,omitempty" yaml:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// Image is a required field but we use omitempty
	// because a empty image value will override a parent's image value with a empty string value
	Image        string `json:"image,omitempty" yaml:"image,omitempty"`
	MemoryLimit  string `json:"memoryLimit,omitempty" yaml:"memoryLimit,omitempty"`
	MountSources bool   `json:"mountSources,omitempty" yaml:"mountSources,omitempty"`

	// Optional specification of the path in the container where project sources should be transferred/mounted when `mountSources` is `true`. When omitted, the value of the `PROJECTS_ROOT` environment variable is used.
	SourceMapping string `json:"sourceMapping,omitempty" yaml:"sourceMapping,omitempty"`

	// List of volumes mounts that should be mounted is this container.
	VolumeMounts []VolumeMount `json:"volumeMounts,omitempty" yaml:"volumeMounts,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

// Endpoint holds information about how an application is exposed
type Endpoint struct {
	Attributes map[string]string `json:"attributes,omitempty" yaml:"attributes,omitempty"`

	// Describes how the endpoint should be exposed on the network. public|internal|none. Default value is "public"
	Exposure ExposureType `json:"exposure,omitempty" yaml:"exposure,omitempty"`

	Path       string `json:"path,omitempty" yaml:"path,omitempty"`
	Secure     bool   `json:"secure,omitempty" yaml:"secure,omitempty"`
	Name       string `json:"name" yaml:"name"`
	TargetPort int32  `json:"targetPort" yaml:"targetPort"`

	// Describes the application and transport protocols of the traffic that will go through this endpoint. Default value is "http"
	Protocol ProtocolType `json:"protocol,omitempty" yaml:"protocol,omitempty"`
}

// Env
type Env struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

// Events Bindings of commands to events. Each command is referred-to by its name.
type DevfileEvents struct {

	// Names of commands that should be executed after the workspace is completely started. In the case of Che-Theia, these commands should be executed after all plugins and extensions have started, including project cloning. This means that those commands are not triggered until the user opens the IDE in his browser.
	PostStart []string `json:"postStart,omitempty" yaml:"postStart,omitempty"`

	// Names of commands that should be executed after stopping the workspace.
	PostStop []string `json:"postStop,omitempty" yaml:"postStop,omitempty"`

	// Names of commands that should be executed before the workspace start. Kubernetes-wise, these commands would typically be executed in init containers of the workspace POD.
	PreStart []string `json:"preStart,omitempty" yaml:"preStart,omitempty"`

	// Names of commands that should be executed before stopping the workspace.
	PreStop []string `json:"preStop,omitempty" yaml:"preStop,omitempty"`
}

// Exec CLI Command executed in a component container
type Exec struct {

	// Optional map of free-form additional command attributes
	Attributes map[string]string `json:"attributes,omitempty" yaml:"attributes,omitempty" patchStrategy:"merge"`

	// The actual command-line string
	CommandLine string `json:"commandLine,omitempty" yaml:"commandLine,omitempty"`

	// Describes component to which given action relates
	Component string `json:"component,omitempty" yaml:"component,omitempty"`

	// Optional list of environment variables that have to be set before running the command
	Env []Env `json:"env,omitempty" yaml:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// Defines the group this command is part of
	Group *Group `json:"group,omitempty" yaml:"group,omitempty"`

	// Optional label that provides a label for this command to be used in Editor UI menus for example
	Label string `json:"label,omitempty" yaml:"label,omitempty"`

	// Working directory where the command should be executed
	WorkingDir string `json:"workingDir,omitempty" yaml:"workingDir,omitempty"`

	// +optional
	// Whether the command is capable to reload itself when source code changes.
	// If set to `true` the command won't be restarted and it is expected to handle file changes on its own.
	//
	// Default value is `false`
	HotReloadCapable bool `json:"hotReloadCapable,omitempty" yaml:"hotReloadCapable,omitempty"`
}

// Composite command containing a list of commands to execute in a component container
type Composite struct {
	// Optional map of free-form additional command attributes
	Attributes map[string]string `json:"attributes,omitempty"`

	// The list of commands to execute in this composite command.
	Commands []string `json:"commands,omitempty"`

	// Defines the group this command is part of
	Group *Group `json:"group,omitempty"`

	// Optional label that provides a label for this command to be used in Editor UI menus for example
	Label string `json:"label,omitempty"`

	// Whether or not the composite command should be executed in parallel
	Parallel bool `json:"parallel,omitempty"`
}

type GitLikeProjectSource struct {
	// The remotes map which should be initialized in the git project. Must have at least one remote configured
	Remotes map[string]string `json:"remotes,omitempty" yaml:"remotes,omitempty"`

	// Defines from what the project should be checked out. Required if there are more than one remote configured
	// +optional
	CheckoutFrom *CheckoutFrom `json:"checkoutFrom,omitempty" yaml:"checkoutFrom,omitempty"`

	// Part of project to populate in the working directory.
	SparseCheckoutDir string `json:"sparseCheckoutDir,omitempty" yaml:"sparseCheckoutDir,omitempty"`
}

// Git Project's Git source
// Github Project's GitHub source
type Git struct {
	GitLikeProjectSource `json:",inline" yaml:",inline"`
}

// Github Project's GitHub source
type Github struct {
	GitLikeProjectSource `json:",inline" yaml:",inline"`
}

// Group Defines the group this command is part of
type Group struct {

	// Identifies the default command for a given group kind
	IsDefault bool `json:"isDefault,omitempty" yaml:"isDefault,omitempty"`

	// Kind of group the command is part of
	Kind DevfileCommandGroupType `json:"kind" yaml:"kind"`
}

// Kubernetes Allows importing into the workspace the Kubernetes resources defined in a given manifest. For example this allows reusing the Kubernetes definitions used to deploy some runtime components in production.
type Kubernetes struct {

	// Inlined manifest
	Inlined string `json:"inlined,omitempty" yaml:"inlined,omitempty"`

	// Mandatory name that allows referencing the component in commands, or inside a parent
	Name string `json:"name" yaml:"name"`

	// Location in a file fetched from a uri.
	Uri string `json:"uri,omitempty" yaml:"uri,omitempty"`
}

// Openshift Configuration overriding for an OpenShift component
type Openshift struct {

	// Inlined manifest
	Inlined string `json:"inlined,omitempty" yaml:"inlined,omitempty"`

	// Mandatory name that allows referencing the component in commands, or inside a parent
	Name string `json:"name" yaml:"name"`

	// Location in a file fetched from a uri.
	Uri string `json:"uri,omitempty" yaml:"uri,omitempty"`
}

// DevfileParent Parent workspace template
type DevfileParent struct {

	// Id in a registry that contains a Devfile yaml file
	Id string `json:"id,omitempty" yaml:"id,omitempty"`

	// Reference to a Kubernetes CRD of type DevWorkspaceTemplate
	Kubernetes Kubernetes `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty"`

	// Uri of a Devfile yaml file
	Uri string `json:"uri,omitempty" yaml:"uri,omitempty"`

	RegistryUrl string `json:"registryUrl,omitempty" yaml:"registryUrl,omitempty"`

	// Projects worked on in the workspace, containing names and sources locations
	Projects []DevfileProject `json:"projects,omitempty" yaml:"projects,omitempty"`

	// Predefined, ready-to-use, workspace-related commands
	Commands []DevfileCommand `json:"commands,omitempty" yaml:"commands,omitempty"`

	// List of the workspace components, such as editor and plugins, user-provided containers, or other types of components
	Components []DevfileComponent `json:"components,omitempty" yaml:"components,omitempty"`

	// StarterProjects is a project that can be used as a starting point when bootstrapping new projects
	StarterProjects []DevfileStarterProject `json:"starterProjects,omitempty" yaml:"starterProjects,omitempty"`
}

// Plugin Allows importing a plugin. Plugins are mainly imported devfiles that contribute components, commands and events as a consistent single unit. They are defined in either YAML files following the devfile syntax, or as `DevWorkspaceTemplate` Kubernetes Custom Resources
type Plugin struct {

	// Overrides of commands encapsulated in a plugin. Overriding is done using a strategic merge
	Commands []*DevfileCommand `json:"commands,omitempty" yaml:"commands,omitempty"`

	// Overrides of components encapsulated in a plugin. Overriding is done using a strategic merge
	Components []*DevfileComponent `json:"components,omitempty" yaml:"components,omitempty"`

	// Id in a registry that contains a Devfile yaml file
	Id string `json:"id,omitempty" yaml:"id,omitempty"`

	// Reference to a Kubernetes CRD of type DevWorkspaceTemplate
	Kubernetes *Kubernetes `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty"`

	// Optional name that allows referencing the component in commands, or inside a parent If omitted it will be infered from the location (uri or registryEntry)
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	RegistryUrl string `json:"registryUrl,omitempty" yaml:"registryUrl,omitempty"`

	// Uri of a Devfile yaml file
	Uri string `json:"uri,omitempty" yaml:"uri,omitempty"`
}

// DevfileProject project defined in devfile
type DevfileProject struct {

	// Path relative to the root of the projects to which this project should be cloned into. This is a unix-style relative path (i.e. uses forward slashes). The path is invalid if it is absolute or tries to escape the project root through the usage of '..'. If not specified, defaults to the project name.
	ClonePath string `json:"clonePath,omitempty" yaml:"clonePath,omitempty"`

	// Project's Git source
	Git *Git `json:"git,omitempty" yaml:"git,omitempty"`

	// Project's GitHub source
	Github *Github `json:"github,omitempty" yaml:"github,omitempty"`

	// Project name
	Name string `json:"name" yaml:"name"`

	// Project's Zip source
	Zip *Zip `json:"zip,omitempty" yaml:"zip,omitempty"`
}

// DevfileStarterProject getting started project
type DevfileStarterProject struct {

	// Project name
	Name string `json:"name" yaml:"name"`

	// Description of a starter project
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Path relative to the root of the projects to which this project should be cloned into. This is a unix-style relative path (i.e. uses forward slashes). The path is invalid if it is absolute or tries to escape the project root through the usage of '..'. If not specified, defaults to the project name.
	ClonePath string `json:"clonePath,omitempty" yaml:"clonePath,omitempty"`

	// Project's Git source
	Git *Git `json:"git,omitempty" yaml:"git,omitempty"`

	// Project's GitHub source
	Github *Github `json:"github,omitempty" yaml:"github,omitempty"`

	// Project's Zip source
	Zip *Zip `json:"zip,omitempty" yaml:"zip,omitempty"`
}

// Volume Allows specifying the definition of a volume shared by several other components
type Volume struct {

	// Size of the volume
	Size string `json:"size,omitempty" yaml:"size,omitempty"`
}

// VolumeMount describes a path where a volume should be mounted to a component container
type VolumeMount struct {

	// The volume mount name is the name of an existing `Volume` component. If several containers mount the same volume name then they will reuse the same volume and will be able to access to the same files.
	Name string `json:"name" yaml:"name"`

	// The path in the component container where the volume should be mounted. If not path is mentioned, default path is the is `/<name>`.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}

// Zip Project's Zip source
type Zip struct {

	// Project's source location address. Should be URL for git and github located projects, or; file:// for zip
	Location string `json:"location,omitempty" yaml:"location,omitempty"`

	// Part of project to populate in the working directory.
	SparseCheckoutDir string `json:"sparseCheckoutDir,omitempty" yaml:"sparseCheckoutDir,omitempty"`
}

// CheckoutFrom Defines from what the project should be checked out. Required if there are more than one remote configured
type CheckoutFrom struct {

	// The remote name should be used as init. Required if there are more than one remote configured
	Remote string `json:"remote,omitempty"`

	// The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found.
	Revision string `json:"revision,omitempty"`
}
