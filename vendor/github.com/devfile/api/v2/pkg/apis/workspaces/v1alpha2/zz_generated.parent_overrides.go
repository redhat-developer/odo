package v1alpha2

import (
	attributes "github.com/devfile/api/v2/pkg/attributes"
)

// +devfile:jsonschema:generate
type ParentOverrides struct {
	OverridesBase `json:",inline"`

	// Overrides of components encapsulated in a parent devfile or a plugin.
	// Overriding is done according to K8S strategic merge patch standard rules.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	// +devfile:toplevellist
	Components []ComponentParentOverride `json:"components,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// Overrides of projects encapsulated in a parent devfile.
	// Overriding is done according to K8S strategic merge patch standard rules.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	// +devfile:toplevellist
	Projects []ProjectParentOverride `json:"projects,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// Overrides of starterProjects encapsulated in a parent devfile.
	// Overriding is done according to K8S strategic merge patch standard rules.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	// +devfile:toplevellist
	StarterProjects []StarterProjectParentOverride `json:"starterProjects,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// Overrides of commands encapsulated in a parent devfile or a plugin.
	// Overriding is done according to K8S strategic merge patch standard rules.
	// +optional
	// +patchMergeKey=id
	// +patchStrategy=merge
	// +devfile:toplevellist
	Commands []CommandParentOverride `json:"commands,omitempty" patchStrategy:"merge" patchMergeKey:"id"`
}

//+k8s:openapi-gen=true
type ComponentParentOverride struct {

	// Mandatory name that allows referencing the component
	// from other elements (such as commands) or from an external
	// devfile that may reference this component through a parent or a plugin.
	// +kubebuilder:validation:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	// +kubebuilder:validation:MaxLength=63
	Name string `json:"name"`

	// Map of implementation-dependant free-form YAML attributes.
	// +optional
	Attributes                   attributes.Attributes `json:"attributes,omitempty"`
	ComponentUnionParentOverride `json:",inline"`
}

type ProjectParentOverride struct {

	// Project name
	// +kubebuilder:validation:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	// +kubebuilder:validation:MaxLength=63
	Name string `json:"name"`

	// Map of implementation-dependant free-form YAML attributes.
	// +optional
	Attributes attributes.Attributes `json:"attributes,omitempty"`

	// Path relative to the root of the projects to which this project should be cloned into. This is a unix-style relative path (i.e. uses forward slashes). The path is invalid if it is absolute or tries to escape the project root through the usage of '..'. If not specified, defaults to the project name.
	// +optional
	ClonePath string `json:"clonePath,omitempty"`

	// Populate the project sparsely with selected directories.
	// +optional
	SparseCheckoutDirs []string `json:"sparseCheckoutDirs,omitempty"`

	ProjectSourceParentOverride `json:",inline"`
}

type StarterProjectParentOverride struct {

	// Project name
	// +kubebuilder:validation:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	// +kubebuilder:validation:MaxLength=63
	Name string `json:"name"`

	// Map of implementation-dependant free-form YAML attributes.
	// +optional
	Attributes attributes.Attributes `json:"attributes,omitempty"`

	// Description of a starter project
	// +optional
	Description string `json:"description,omitempty"`

	// Sub-directory from a starter project to be used as root for starter project.
	// +optional
	SubDir string `json:"subDir,omitempty"`

	ProjectSourceParentOverride `json:",inline"`
}

type CommandParentOverride struct {

	// Mandatory identifier that allows referencing
	// this command in composite commands, from
	// a parent, or in events.
	// +kubebuilder:validation:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	// +kubebuilder:validation:MaxLength=63
	Id string `json:"id"`

	// Map of implementation-dependant free-form YAML attributes.
	// +optional
	Attributes                 attributes.Attributes `json:"attributes,omitempty"`
	CommandUnionParentOverride `json:",inline"`
}

// +union
type ComponentUnionParentOverride struct {

	// +kubebuilder:validation:Enum=Container;Kubernetes;Openshift;Volume;Plugin
	// Type of component
	//
	// +unionDiscriminator
	// +optional
	ComponentType ComponentTypeParentOverride `json:"componentType,omitempty"`

	// Allows adding and configuring workspace-related containers
	// +optional
	Container *ContainerComponentParentOverride `json:"container,omitempty"`

	// Allows importing into the workspace the Kubernetes resources
	// defined in a given manifest. For example this allows reusing the Kubernetes
	// definitions used to deploy some runtime components in production.
	//
	// +optional
	Kubernetes *KubernetesComponentParentOverride `json:"kubernetes,omitempty"`

	// Allows importing into the workspace the OpenShift resources
	// defined in a given manifest. For example this allows reusing the OpenShift
	// definitions used to deploy some runtime components in production.
	//
	// +optional
	Openshift *OpenshiftComponentParentOverride `json:"openshift,omitempty"`

	// Allows specifying the definition of a volume
	// shared by several other components
	// +optional
	Volume *VolumeComponentParentOverride `json:"volume,omitempty"`

	// Allows importing a plugin.
	//
	// Plugins are mainly imported devfiles that contribute components, commands
	// and events as a consistent single unit. They are defined in either YAML files
	// following the devfile syntax,
	// or as `DevWorkspaceTemplate` Kubernetes Custom Resources
	// +optional
	// +devfile:overrides:include:omitInPlugin=true
	Plugin *PluginComponentParentOverride `json:"plugin,omitempty"`
}

// +union
type ProjectSourceParentOverride struct {

	// +kubebuilder:validation:Enum=Git;Github;Zip
	// Type of project source
	// +
	// +unionDiscriminator
	// +optional
	SourceType ProjectSourceTypeParentOverride `json:"sourceType,omitempty"`

	// Project's Git source
	// +optional
	Git *GitProjectSourceParentOverride `json:"git,omitempty"`

	// Project's GitHub source
	// +optional
	Github *GithubProjectSourceParentOverride `json:"github,omitempty"`

	// Project's Zip source
	// +optional
	Zip *ZipProjectSourceParentOverride `json:"zip,omitempty"`
}

// +union
type CommandUnionParentOverride struct {

	// +kubebuilder:validation:Enum=Exec;Apply;VscodeTask;VscodeLaunch;Composite
	// Type of workspace command
	// +unionDiscriminator
	// +optional
	CommandType CommandTypeParentOverride `json:"commandType,omitempty"`

	// CLI Command executed in an existing component container
	// +optional
	Exec *ExecCommandParentOverride `json:"exec,omitempty"`

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
	Apply *ApplyCommandParentOverride `json:"apply,omitempty"`

	// Command providing the definition of a VsCode Task
	// +optional
	VscodeTask *VscodeConfigurationCommandParentOverride `json:"vscodeTask,omitempty"`

	// Command providing the definition of a VsCode launch action
	// +optional
	VscodeLaunch *VscodeConfigurationCommandParentOverride `json:"vscodeLaunch,omitempty"`

	// Composite command that allows executing several sub-commands
	// either sequentially or concurrently
	// +optional
	Composite *CompositeCommandParentOverride `json:"composite,omitempty"`
}

// ComponentType describes the type of component.
// Only one of the following component type may be specified.
type ComponentTypeParentOverride string

// Component that allows the developer to add a configured container into his workspace
type ContainerComponentParentOverride struct {
	BaseComponentParentOverride `json:",inline"`
	ContainerParentOverride     `json:",inline"`
	Endpoints                   []EndpointParentOverride `json:"endpoints,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

// Component that allows partly importing Kubernetes resources into the workspace POD
type KubernetesComponentParentOverride struct {
	K8sLikeComponentParentOverride `json:",inline"`
}

// Component that allows partly importing Openshift resources into the workspace POD
type OpenshiftComponentParentOverride struct {
	K8sLikeComponentParentOverride `json:",inline"`
}

// Component that allows the developer to declare and configure a volume into his workspace
type VolumeComponentParentOverride struct {
	BaseComponentParentOverride `json:",inline"`
	VolumeParentOverride        `json:",inline"`
}

type PluginComponentParentOverride struct {
	BaseComponentParentOverride   `json:",inline"`
	ImportReferenceParentOverride `json:",inline"`
	PluginOverridesParentOverride `json:",inline"`
}

// ProjectSourceType describes the type of Project sources.
// Only one of the following project sources may be specified.
// If none of the following policies is specified, the default one
// is AllowConcurrent.
type ProjectSourceTypeParentOverride string

type GitProjectSourceParentOverride struct {
	GitLikeProjectSourceParentOverride `json:",inline"`
}

type GithubProjectSourceParentOverride struct {
	GitLikeProjectSourceParentOverride `json:",inline"`
}

type ZipProjectSourceParentOverride struct {
	CommonProjectSourceParentOverride `json:",inline"`

	// Zip project's source location address. Should be file path of the archive, e.g. file://$FILE_PATH
	// +required
	Location string `json:"location,omitempty"`
}

// CommandType describes the type of command.
// Only one of the following command type may be specified.
type CommandTypeParentOverride string

type ExecCommandParentOverride struct {
	LabeledCommandParentOverride `json:",inline"`

	//  +optional
	// The actual command-line string
	//
	// Special variables that can be used:
	//
	//  - `$PROJECTS_ROOT`: A path where projects sources are mounted as defined by container component's sourceMapping.
	//
	//  - `$PROJECT_SOURCE`: A path to a project source ($PROJECTS_ROOT/<project-name>). If there are multiple projects, this will point to the directory of the first one.
	CommandLine string `json:"commandLine,omitempty"`

	//  +optional
	// Describes component to which given action relates
	//
	Component string `json:"component,omitempty"`

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
	Env []EnvVarParentOverride `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// +optional
	// Whether the command is capable to reload itself when source code changes.
	// If set to `true` the command won't be restarted and it is expected to handle file changes on its own.
	//
	// Default value is `false`
	HotReloadCapable bool `json:"hotReloadCapable,omitempty"`
}

type ApplyCommandParentOverride struct {
	LabeledCommandParentOverride `json:",inline"`

	//  +optional
	// Describes component that will be applied
	//
	Component string `json:"component,omitempty"`
}

type VscodeConfigurationCommandParentOverride struct {
	BaseCommandParentOverride                        `json:",inline"`
	VscodeConfigurationCommandLocationParentOverride `json:",inline"`
}

type CompositeCommandParentOverride struct {
	LabeledCommandParentOverride `json:",inline"`

	// The commands that comprise this composite command
	Commands []string `json:"commands,omitempty" patchStrategy:"replace"`

	// Indicates if the sub-commands should be executed concurrently
	// +optional
	Parallel bool `json:"parallel,omitempty"`
}

// Workspace component: Anything that will bring additional features / tooling / behaviour / context
// to the workspace, in order to make working in it easier.
type BaseComponentParentOverride struct {
}

type ContainerParentOverride struct {
	//  +optional
	Image string `json:"image,omitempty"`

	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	// Environment variables used in this container.
	//
	// The following variables are reserved and cannot be overridden via env:
	//
	//  - `$PROJECTS_ROOT`
	//
	//  - `$PROJECT_SOURCE`
	Env []EnvVarParentOverride `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// +optional
	// List of volumes mounts that should be mounted is this container.
	VolumeMounts []VolumeMountParentOverride `json:"volumeMounts,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// +optional
	MemoryLimit string `json:"memoryLimit,omitempty"`

	// +optional
	MemoryRequest string `json:"memoryRequest,omitempty"`

	// +optional
	CpuLimit string `json:"cpuLimit,omitempty"`

	// +optional
	CpuRequest string `json:"cpuRequest,omitempty"`

	// The command to run in the dockerimage component instead of the default one provided in the image.
	//
	// Defaults to an empty array, meaning use whatever is defined in the image.
	// +optional
	Command []string `json:"command,omitempty" patchStrategy:"replace"`

	// The arguments to supply to the command running the dockerimage component. The arguments are supplied either to the default command provided in the image or to the overridden command.
	//
	// Defaults to an empty array, meaning use whatever is defined in the image.
	// +optional
	Args []string `json:"args,omitempty" patchStrategy:"replace"`

	// Toggles whether or not the project source code should
	// be mounted in the component.
	//
	// Defaults to true for all component types except plugins and components that set `dedicatedPod` to true.
	// +optional
	MountSources *bool `json:"mountSources,omitempty"`

	// Optional specification of the path in the container where
	// project sources should be transferred/mounted when `mountSources` is `true`.
	// When omitted, the default value of /projects is used.
	// +optional
	SourceMapping string `json:"sourceMapping,omitempty"`

	// Specify if a container should run in its own separated pod,
	// instead of running as part of the main development environment pod.
	//
	// Default value is `false`
	// +optional
	DedicatedPod bool `json:"dedicatedPod,omitempty"`
}

type EndpointParentOverride struct {

	// +kubebuilder:validation:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	// +kubebuilder:validation:MaxLength=63
	Name string `json:"name"`

	//  +optional
	TargetPort int `json:"targetPort,omitempty"`

	// Describes how the endpoint should be exposed on the network.
	//
	// - `public` means that the endpoint will be exposed on the public network, typically through
	// a K8S ingress or an OpenShift route.
	//
	// - `internal` means that the endpoint will be exposed internally outside of the main workspace POD,
	// typically by K8S services, to be consumed by other elements running
	// on the same cloud internal network.
	//
	// - `none` means that the endpoint will not be exposed and will only be accessible
	// inside the main workspace POD, on a local address.
	//
	// Default value is `public`
	// +optional
	Exposure EndpointExposureParentOverride `json:"exposure,omitempty"`

	// Describes the application and transport protocols of the traffic that will go through this endpoint.
	//
	// - `http`: Endpoint will have `http` traffic, typically on a TCP connection.
	// It will be automaticaly promoted to `https` when the `secure` field is set to `true`.
	//
	// - `https`: Endpoint will have `https` traffic, typically on a TCP connection.
	//
	// - `ws`: Endpoint will have `ws` traffic, typically on a TCP connection.
	// It will be automaticaly promoted to `wss` when the `secure` field is set to `true`.
	//
	// - `wss`: Endpoint will have `wss` traffic, typically on a TCP connection.
	//
	// - `tcp`: Endpoint will have traffic on a TCP connection, without specifying an application protocol.
	//
	// - `udp`: Endpoint will have traffic on an UDP connection, without specifying an application protocol.
	//
	// Default value is `http`
	// +optional
	Protocol EndpointProtocolParentOverride `json:"protocol,omitempty"`

	// Describes whether the endpoint should be secured and protected by some
	// authentication process. This requires a protocol of `https` or `wss`.
	// +optional
	Secure bool `json:"secure,omitempty"`

	// Path of the endpoint URL
	// +optional
	Path string `json:"path,omitempty"`

	// Map of implementation-dependant string-based free-form attributes.
	//
	// Examples of Che-specific attributes:
	//
	// - cookiesAuthEnabled: "true" / "false",
	//
	// - type: "terminal" / "ide" / "ide-dev",
	// +optional
	Attributes attributes.Attributes `json:"attributes,omitempty"`
}

type K8sLikeComponentParentOverride struct {
	BaseComponentParentOverride            `json:",inline"`
	K8sLikeComponentLocationParentOverride `json:",inline"`
	Endpoints                              []EndpointParentOverride `json:"endpoints,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

// Volume that should be mounted to a component container
type VolumeParentOverride struct {

	// +optional
	// Size of the volume
	Size string `json:"size,omitempty"`
}

type ImportReferenceParentOverride struct {
	ImportReferenceUnionParentOverride `json:",inline"`

	// +optional
	RegistryUrl string `json:"registryUrl,omitempty"`
}

type PluginOverridesParentOverride struct {
	OverridesBaseParentOverride `json:",inline"`

	// Overrides of components encapsulated in a parent devfile or a plugin.
	// Overriding is done according to K8S strategic merge patch standard rules.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	// +devfile:toplevellist
	Components []ComponentPluginOverrideParentOverride `json:"components,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// Overrides of commands encapsulated in a parent devfile or a plugin.
	// Overriding is done according to K8S strategic merge patch standard rules.
	// +optional
	// +patchMergeKey=id
	// +patchStrategy=merge
	// +devfile:toplevellist
	Commands []CommandPluginOverrideParentOverride `json:"commands,omitempty" patchStrategy:"merge" patchMergeKey:"id"`
}

type GitLikeProjectSourceParentOverride struct {
	CommonProjectSourceParentOverride `json:",inline"`

	// Defines from what the project should be checked out. Required if there are more than one remote configured
	// +optional
	CheckoutFrom *CheckoutFromParentOverride `json:"checkoutFrom,omitempty"`

	//  +optional
	// The remotes map which should be initialized in the git project. Must have at least one remote configured
	Remotes map[string]string `json:"remotes,omitempty"`
}

type CommonProjectSourceParentOverride struct {
}

type LabeledCommandParentOverride struct {
	BaseCommandParentOverride `json:",inline"`

	// +optional
	// Optional label that provides a label for this command
	// to be used in Editor UI menus for example
	Label string `json:"label,omitempty"`
}

type EnvVarParentOverride struct {
	Name string `json:"name" yaml:"name"`
	//  +optional
	Value string `json:"value,omitempty" yaml:"value"`
}

type BaseCommandParentOverride struct {

	// +optional
	// Defines the group this command is part of
	Group *CommandGroupParentOverride `json:"group,omitempty"`
}

// +union
type VscodeConfigurationCommandLocationParentOverride struct {

	// +kubebuilder:validation:Enum=Uri;Inlined
	// Type of Vscode configuration command location
	// +
	// +unionDiscriminator
	// +optional
	LocationType VscodeConfigurationCommandLocationTypeParentOverride `json:"locationType,omitempty"`

	// Location as an absolute of relative URI
	// the VsCode configuration will be fetched from
	// +optional
	Uri string `json:"uri,omitempty"`

	// Inlined content of the VsCode configuration
	// +optional
	Inlined string `json:"inlined,omitempty"`
}

// Volume that should be mounted to a component container
type VolumeMountParentOverride struct {

	// The volume mount name is the name of an existing `Volume` component.
	// If several containers mount the same volume name
	// then they will reuse the same volume and will be able to access to the same files.
	// +kubebuilder:validation:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	// +kubebuilder:validation:MaxLength=63
	Name string `json:"name"`

	// The path in the component container where the volume should be mounted.
	// If not path is mentioned, default path is the is `/<name>`.
	// +optional
	Path string `json:"path,omitempty"`
}

// EndpointExposure describes the way an endpoint is exposed on the network.
// Only one of the following exposures may be specified: public, internal, none.
// +kubebuilder:validation:Enum=public;internal;none
type EndpointExposureParentOverride string

// EndpointProtocol defines the application and transport protocols of the traffic that will go through this endpoint.
// Only one of the following protocols may be specified: http, ws, tcp, udp.
// +kubebuilder:validation:Enum=http;https;ws;wss;tcp;udp
type EndpointProtocolParentOverride string

// +union
type K8sLikeComponentLocationParentOverride struct {

	// +kubebuilder:validation:Enum=Uri;Inlined
	// Type of Kubernetes-like location
	// +
	// +unionDiscriminator
	// +optional
	LocationType K8sLikeComponentLocationTypeParentOverride `json:"locationType,omitempty"`

	// Location in a file fetched from a uri.
	// +optional
	Uri string `json:"uri,omitempty"`

	// Inlined manifest
	// +optional
	Inlined string `json:"inlined,omitempty"`
}

// Location from where the an import reference is retrieved
// +union
type ImportReferenceUnionParentOverride struct {

	// +kubebuilder:validation:Enum=Uri;Id;Kubernetes
	// type of location from where the referenced template structure should be retrieved
	// +
	// +unionDiscriminator
	// +optional
	ImportReferenceType ImportReferenceTypeParentOverride `json:"importReferenceType,omitempty"`

	// Uri of a Devfile yaml file
	// +optional
	Uri string `json:"uri,omitempty"`

	// Id in a registry that contains a Devfile yaml file
	// +optional
	Id string `json:"id,omitempty"`

	// Reference to a Kubernetes CRD of type DevWorkspaceTemplate
	// +optional
	Kubernetes *KubernetesCustomResourceImportReferenceParentOverride `json:"kubernetes,omitempty"`
}

// OverridesBase is used in the Overrides generator in order to provide a common base for the generated Overrides
// So please be careful when renaming
type OverridesBaseParentOverride struct{}

//+k8s:openapi-gen=true
type ComponentPluginOverrideParentOverride struct {

	// Mandatory name that allows referencing the component
	// from other elements (such as commands) or from an external
	// devfile that may reference this component through a parent or a plugin.
	// +kubebuilder:validation:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	// +kubebuilder:validation:MaxLength=63
	Name string `json:"name"`

	// Map of implementation-dependant free-form YAML attributes.
	// +optional
	Attributes                                 attributes.Attributes `json:"attributes,omitempty"`
	ComponentUnionPluginOverrideParentOverride `json:",inline"`
}

type CommandPluginOverrideParentOverride struct {

	// Mandatory identifier that allows referencing
	// this command in composite commands, from
	// a parent, or in events.
	// +kubebuilder:validation:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	// +kubebuilder:validation:MaxLength=63
	Id string `json:"id"`

	// Map of implementation-dependant free-form YAML attributes.
	// +optional
	Attributes                               attributes.Attributes `json:"attributes,omitempty"`
	CommandUnionPluginOverrideParentOverride `json:",inline"`
}

type CheckoutFromParentOverride struct {

	// The revision to checkout from. Should be branch name, tag or commit id.
	// Default branch is used if missing or specified revision is not found.
	// +optional
	Revision string `json:"revision,omitempty"`

	// The remote name should be used as init. Required if there are more than one remote configured
	// +optional
	Remote string `json:"remote,omitempty"`
}

type CommandGroupParentOverride struct {

	//  +optional
	// Kind of group the command is part of
	Kind CommandGroupKindParentOverride `json:"kind,omitempty"`

	// +optional
	// Identifies the default command for a given group kind
	IsDefault bool `json:"isDefault,omitempty"`
}

// VscodeConfigurationCommandLocationType describes the type of
// the location the configuration is fetched from.
// Only one of the following component type may be specified.
type VscodeConfigurationCommandLocationTypeParentOverride string

// K8sLikeComponentLocationType describes the type of
// the location the configuration is fetched from.
// Only one of the following component type may be specified.
type K8sLikeComponentLocationTypeParentOverride string

// ImportReferenceType describes the type of location
// from where the referenced template structure should be retrieved.
// Only one of the following parent locations may be specified.
type ImportReferenceTypeParentOverride string

type KubernetesCustomResourceImportReferenceParentOverride struct {
	//  +optional
	Name string `json:"name,omitempty"`

	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// +union
type ComponentUnionPluginOverrideParentOverride struct {

	// +kubebuilder:validation:Enum=Container;Kubernetes;Openshift;Volume
	// Type of component
	//
	// +unionDiscriminator
	// +optional
	ComponentType ComponentTypePluginOverrideParentOverride `json:"componentType,omitempty"`

	// Allows adding and configuring workspace-related containers
	// +optional
	Container *ContainerComponentPluginOverrideParentOverride `json:"container,omitempty"`

	// Allows importing into the workspace the Kubernetes resources
	// defined in a given manifest. For example this allows reusing the Kubernetes
	// definitions used to deploy some runtime components in production.
	//
	// +optional
	Kubernetes *KubernetesComponentPluginOverrideParentOverride `json:"kubernetes,omitempty"`

	// Allows importing into the workspace the OpenShift resources
	// defined in a given manifest. For example this allows reusing the OpenShift
	// definitions used to deploy some runtime components in production.
	//
	// +optional
	Openshift *OpenshiftComponentPluginOverrideParentOverride `json:"openshift,omitempty"`

	// Allows specifying the definition of a volume
	// shared by several other components
	// +optional
	Volume *VolumeComponentPluginOverrideParentOverride `json:"volume,omitempty"`
}

// +union
type CommandUnionPluginOverrideParentOverride struct {

	// +kubebuilder:validation:Enum=Exec;Apply;VscodeTask;VscodeLaunch;Composite
	// Type of workspace command
	// +unionDiscriminator
	// +optional
	CommandType CommandTypePluginOverrideParentOverride `json:"commandType,omitempty"`

	// CLI Command executed in an existing component container
	// +optional
	Exec *ExecCommandPluginOverrideParentOverride `json:"exec,omitempty"`

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
	Apply *ApplyCommandPluginOverrideParentOverride `json:"apply,omitempty"`

	// Command providing the definition of a VsCode Task
	// +optional
	VscodeTask *VscodeConfigurationCommandPluginOverrideParentOverride `json:"vscodeTask,omitempty"`

	// Command providing the definition of a VsCode launch action
	// +optional
	VscodeLaunch *VscodeConfigurationCommandPluginOverrideParentOverride `json:"vscodeLaunch,omitempty"`

	// Composite command that allows executing several sub-commands
	// either sequentially or concurrently
	// +optional
	Composite *CompositeCommandPluginOverrideParentOverride `json:"composite,omitempty"`
}

// CommandGroupKind describes the kind of command group.
// +kubebuilder:validation:Enum=build;run;test;debug
type CommandGroupKindParentOverride string

// ComponentType describes the type of component.
// Only one of the following component type may be specified.
type ComponentTypePluginOverrideParentOverride string

// Component that allows the developer to add a configured container into his workspace
type ContainerComponentPluginOverrideParentOverride struct {
	BaseComponentPluginOverrideParentOverride `json:",inline"`
	ContainerPluginOverrideParentOverride     `json:",inline"`
	Endpoints                                 []EndpointPluginOverrideParentOverride `json:"endpoints,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

// Component that allows partly importing Kubernetes resources into the workspace POD
type KubernetesComponentPluginOverrideParentOverride struct {
	K8sLikeComponentPluginOverrideParentOverride `json:",inline"`
}

// Component that allows partly importing Openshift resources into the workspace POD
type OpenshiftComponentPluginOverrideParentOverride struct {
	K8sLikeComponentPluginOverrideParentOverride `json:",inline"`
}

// Component that allows the developer to declare and configure a volume into his workspace
type VolumeComponentPluginOverrideParentOverride struct {
	BaseComponentPluginOverrideParentOverride `json:",inline"`
	VolumePluginOverrideParentOverride        `json:",inline"`
}

// CommandType describes the type of command.
// Only one of the following command type may be specified.
type CommandTypePluginOverrideParentOverride string

type ExecCommandPluginOverrideParentOverride struct {
	LabeledCommandPluginOverrideParentOverride `json:",inline"`

	//  +optional
	// The actual command-line string
	//
	// Special variables that can be used:
	//
	//  - `$PROJECTS_ROOT`: A path where projects sources are mounted as defined by container component's sourceMapping.
	//
	//  - `$PROJECT_SOURCE`: A path to a project source ($PROJECTS_ROOT/<project-name>). If there are multiple projects, this will point to the directory of the first one.
	CommandLine string `json:"commandLine,omitempty"`

	//  +optional
	// Describes component to which given action relates
	//
	Component string `json:"component,omitempty"`

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
	Env []EnvVarPluginOverrideParentOverride `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// +optional
	// Whether the command is capable to reload itself when source code changes.
	// If set to `true` the command won't be restarted and it is expected to handle file changes on its own.
	//
	// Default value is `false`
	HotReloadCapable bool `json:"hotReloadCapable,omitempty"`
}

type ApplyCommandPluginOverrideParentOverride struct {
	LabeledCommandPluginOverrideParentOverride `json:",inline"`

	//  +optional
	// Describes component that will be applied
	//
	Component string `json:"component,omitempty"`
}

type VscodeConfigurationCommandPluginOverrideParentOverride struct {
	BaseCommandPluginOverrideParentOverride                        `json:",inline"`
	VscodeConfigurationCommandLocationPluginOverrideParentOverride `json:",inline"`
}

type CompositeCommandPluginOverrideParentOverride struct {
	LabeledCommandPluginOverrideParentOverride `json:",inline"`

	// The commands that comprise this composite command
	Commands []string `json:"commands,omitempty" patchStrategy:"replace"`

	// Indicates if the sub-commands should be executed concurrently
	// +optional
	Parallel bool `json:"parallel,omitempty"`
}

// Workspace component: Anything that will bring additional features / tooling / behaviour / context
// to the workspace, in order to make working in it easier.
type BaseComponentPluginOverrideParentOverride struct {
}

type ContainerPluginOverrideParentOverride struct {

	//  +optional
	Image string `json:"image,omitempty"`

	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	// Environment variables used in this container.
	//
	// The following variables are reserved and cannot be overridden via env:
	//
	//  - `$PROJECTS_ROOT`
	//
	//  - `$PROJECT_SOURCE`
	Env []EnvVarPluginOverrideParentOverride `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// +optional
	// List of volumes mounts that should be mounted is this container.
	VolumeMounts []VolumeMountPluginOverrideParentOverride `json:"volumeMounts,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// +optional
	MemoryLimit string `json:"memoryLimit,omitempty"`

	// +optional
	MemoryRequest string `json:"memoryRequest,omitempty"`

	// +optional
	CpuLimit string `json:"cpuLimit,omitempty"`

	// +optional
	CpuRequest string `json:"cpuRequest,omitempty"`

	// The command to run in the dockerimage component instead of the default one provided in the image.
	//
	// Defaults to an empty array, meaning use whatever is defined in the image.
	// +optional
	Command []string `json:"command,omitempty" patchStrategy:"replace"`

	// The arguments to supply to the command running the dockerimage component. The arguments are supplied either to the default command provided in the image or to the overridden command.
	//
	// Defaults to an empty array, meaning use whatever is defined in the image.
	// +optional
	Args []string `json:"args,omitempty" patchStrategy:"replace"`

	// Toggles whether or not the project source code should
	// be mounted in the component.
	//
	// Defaults to true for all component types except plugins and components that set `dedicatedPod` to true.
	// +optional
	MountSources *bool `json:"mountSources,omitempty"`

	// Optional specification of the path in the container where
	// project sources should be transferred/mounted when `mountSources` is `true`.
	// When omitted, the default value of /projects is used.
	// +optional
	SourceMapping string `json:"sourceMapping,omitempty"`

	// Specify if a container should run in its own separated pod,
	// instead of running as part of the main development environment pod.
	//
	// Default value is `false`
	// +optional
	DedicatedPod bool `json:"dedicatedPod,omitempty"`
}

type EndpointPluginOverrideParentOverride struct {

	// +kubebuilder:validation:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	// +kubebuilder:validation:MaxLength=63
	Name string `json:"name"`

	//  +optional
	TargetPort int `json:"targetPort,omitempty"`

	// Describes how the endpoint should be exposed on the network.
	//
	// - `public` means that the endpoint will be exposed on the public network, typically through
	// a K8S ingress or an OpenShift route.
	//
	// - `internal` means that the endpoint will be exposed internally outside of the main workspace POD,
	// typically by K8S services, to be consumed by other elements running
	// on the same cloud internal network.
	//
	// - `none` means that the endpoint will not be exposed and will only be accessible
	// inside the main workspace POD, on a local address.
	//
	// Default value is `public`
	// +optional
	Exposure EndpointExposurePluginOverrideParentOverride `json:"exposure,omitempty"`

	// Describes the application and transport protocols of the traffic that will go through this endpoint.
	//
	// - `http`: Endpoint will have `http` traffic, typically on a TCP connection.
	// It will be automaticaly promoted to `https` when the `secure` field is set to `true`.
	//
	// - `https`: Endpoint will have `https` traffic, typically on a TCP connection.
	//
	// - `ws`: Endpoint will have `ws` traffic, typically on a TCP connection.
	// It will be automaticaly promoted to `wss` when the `secure` field is set to `true`.
	//
	// - `wss`: Endpoint will have `wss` traffic, typically on a TCP connection.
	//
	// - `tcp`: Endpoint will have traffic on a TCP connection, without specifying an application protocol.
	//
	// - `udp`: Endpoint will have traffic on an UDP connection, without specifying an application protocol.
	//
	// Default value is `http`
	// +optional
	Protocol EndpointProtocolPluginOverrideParentOverride `json:"protocol,omitempty"`

	// Describes whether the endpoint should be secured and protected by some
	// authentication process. This requires a protocol of `https` or `wss`.
	// +optional
	Secure bool `json:"secure,omitempty"`

	// Path of the endpoint URL
	// +optional
	Path string `json:"path,omitempty"`

	// Map of implementation-dependant string-based free-form attributes.
	//
	// Examples of Che-specific attributes:
	//
	// - cookiesAuthEnabled: "true" / "false",
	//
	// - type: "terminal" / "ide" / "ide-dev",
	// +optional
	Attributes attributes.Attributes `json:"attributes,omitempty"`
}

type K8sLikeComponentPluginOverrideParentOverride struct {
	BaseComponentPluginOverrideParentOverride            `json:",inline"`
	K8sLikeComponentLocationPluginOverrideParentOverride `json:",inline"`
	Endpoints                                            []EndpointPluginOverrideParentOverride `json:"endpoints,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

// Volume that should be mounted to a component container
type VolumePluginOverrideParentOverride struct {

	// +optional
	// Size of the volume
	Size string `json:"size,omitempty"`
}

type LabeledCommandPluginOverrideParentOverride struct {
	BaseCommandPluginOverrideParentOverride `json:",inline"`

	// +optional
	// Optional label that provides a label for this command
	// to be used in Editor UI menus for example
	Label string `json:"label,omitempty"`
}

type EnvVarPluginOverrideParentOverride struct {
	Name string `json:"name" yaml:"name"`

	//  +optional
	Value string `json:"value,omitempty" yaml:"value"`
}

type BaseCommandPluginOverrideParentOverride struct {

	// +optional
	// Defines the group this command is part of
	Group *CommandGroupPluginOverrideParentOverride `json:"group,omitempty"`
}

// +union
type VscodeConfigurationCommandLocationPluginOverrideParentOverride struct {

	// +kubebuilder:validation:Enum=Uri;Inlined
	// Type of Vscode configuration command location
	// +
	// +unionDiscriminator
	// +optional
	LocationType VscodeConfigurationCommandLocationTypePluginOverrideParentOverride `json:"locationType,omitempty"`

	// Location as an absolute of relative URI
	// the VsCode configuration will be fetched from
	// +optional
	Uri string `json:"uri,omitempty"`

	// Inlined content of the VsCode configuration
	// +optional
	Inlined string `json:"inlined,omitempty"`
}

// Volume that should be mounted to a component container
type VolumeMountPluginOverrideParentOverride struct {

	// The volume mount name is the name of an existing `Volume` component.
	// If several containers mount the same volume name
	// then they will reuse the same volume and will be able to access to the same files.
	// +kubebuilder:validation:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	// +kubebuilder:validation:MaxLength=63
	Name string `json:"name"`

	// The path in the component container where the volume should be mounted.
	// If not path is mentioned, default path is the is `/<name>`.
	// +optional
	Path string `json:"path,omitempty"`
}

// EndpointExposure describes the way an endpoint is exposed on the network.
// Only one of the following exposures may be specified: public, internal, none.
// +kubebuilder:validation:Enum=public;internal;none
type EndpointExposurePluginOverrideParentOverride string

// EndpointProtocol defines the application and transport protocols of the traffic that will go through this endpoint.
// Only one of the following protocols may be specified: http, ws, tcp, udp.
// +kubebuilder:validation:Enum=http;https;ws;wss;tcp;udp
type EndpointProtocolPluginOverrideParentOverride string

// +union
type K8sLikeComponentLocationPluginOverrideParentOverride struct {

	// +kubebuilder:validation:Enum=Uri;Inlined
	// Type of Kubernetes-like location
	// +
	// +unionDiscriminator
	// +optional
	LocationType K8sLikeComponentLocationTypePluginOverrideParentOverride `json:"locationType,omitempty"`

	// Location in a file fetched from a uri.
	// +optional
	Uri string `json:"uri,omitempty"`

	// Inlined manifest
	// +optional
	Inlined string `json:"inlined,omitempty"`
}

type CommandGroupPluginOverrideParentOverride struct {

	//  +optional
	// Kind of group the command is part of
	Kind CommandGroupKindPluginOverrideParentOverride `json:"kind,omitempty"`

	// +optional
	// Identifies the default command for a given group kind
	IsDefault bool `json:"isDefault,omitempty"`
}

// VscodeConfigurationCommandLocationType describes the type of
// the location the configuration is fetched from.
// Only one of the following component type may be specified.
type VscodeConfigurationCommandLocationTypePluginOverrideParentOverride string

// K8sLikeComponentLocationType describes the type of
// the location the configuration is fetched from.
// Only one of the following component type may be specified.
type K8sLikeComponentLocationTypePluginOverrideParentOverride string

// CommandGroupKind describes the kind of command group.
// +kubebuilder:validation:Enum=build;run;test;debug
type CommandGroupKindPluginOverrideParentOverride string

func (overrides ParentOverrides) isOverride() {}
