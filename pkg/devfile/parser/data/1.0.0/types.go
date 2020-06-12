package version100

// Devfile100 struct maps to devfile 1.0.0 version schema
type Devfile100 struct {

	// Devfile section "apiVersion"
	ApiVersion ApiVersion `yaml:"apiVersion" json:"apiVersion"`

	// Devfile section "metadata"
	Metadata Metadata `yaml:"metadata" json:"metadata"`

	// Devfile section projects
	Projects []Project `yaml:"projects,omitempty" json:"projects,omitempty"`

	Attributes Attributes `yaml:"attributes,omitempty" json:"attributes,omitempty"`

	// Description of the workspace components, such as editor and plugins
	Components []Component `yaml:"components,omitempty" json:"components,omitempty"`

	// Description of the predefined commands to be available in workspace
	Commands []Command `yaml:"commands,omitempty" json:"commands,omitempty"`
}

// -------------- Supported devfile project types ------------ //
// DevfileProjectType store valid devfile project types
type ProjectType string

const (
	ProjectTypeGit    ProjectType = "git"
	ProjectTypeGitHub ProjectType = "github"
	ProjectTypeZip    ProjectType = "zip"
)

var SupportedProjectTypes = []ProjectType{ProjectTypeGit}

// -------------- Supported devfile component types ------------ //
// DevfileComponentType stores valid devfile component types
type ComponentType string

const (
	DevfileComponentTypeCheEditor   ComponentType = "cheEditor"
	DevfileComponentTypeChePlugin   ComponentType = "chePlugin"
	DevfileComponentTypeDockerimage ComponentType = "dockerimage"
	DevfileComponentTypeKubernetes  ComponentType = "kubernetes"
	DevfileComponentTypeOpenshift   ComponentType = "openshift"
)

// -------------- Supported devfile command types ------------ //
type CommandType string

const (
	DevfileCommandTypeInit  CommandType = "init"
	DevfileCommandTypeBuild CommandType = "build"
	DevfileCommandTypeRun   CommandType = "run"
	DevfileCommandTypeDebug CommandType = "debug"
	DevfileCommandTypeExec  CommandType = "exec"
)

// ----------- Devfile Schema ---------- //
type Attributes map[string]string

type ApiVersion string

type Metadata struct {

	// Workspaces created from devfile, will use it as base and append random suffix.
	// It's used when name is not defined.
	GenerateName string `yaml:"generateName,omitempty" json:"generateName,omitempty"`

	// The name of the devfile. Workspaces created from devfile, will inherit this
	// name
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
}

// Description of the projects, containing names and sources locations
type Project struct {

	// The path relative to the root of the projects to which this project should be cloned into. This is a unix-style relative path (i.e. uses forward slashes). The path is invalid if it is absolute or tries to escape the project root through the usage of '..'. If not specified, defaults to the project name."
	ClonePath string `yaml:"clonePath,omitempty" json:"clonePath,omitempty"`

	// The Project Name
	Name string `yaml:"name" json:"name"`

	// Describes the project's source - type and location
	Source ProjectSource `yaml:"source" json:"source"`
}

type ProjectSource struct {
	Type ProjectType `yaml:"type" json:"type"`

	// Project's source location address. Should be URL for git and github located projects"
	Location string `yaml:"location" json:"location"`

	// The name of the of the branch to check out after obtaining the source from the location.
	//  The branch has to already exist in the source otherwise the default branch is used.
	//  In case of git, this is also the name of the remote branch to push to.
	Branch string `yaml:"branch,omitempty" json:"branch,omitempty"`

	// The id of the commit to reset the checked out branch to.
	//  Note that this is equivalent to 'startPoint' and provided for convenience.
	CommitId string `yaml:"commitId,omitempty" json:"commitId,omitempty"`

	// Part of project to populate in the working directory.
	SparseCheckoutDir string `yaml:"sparseCheckoutDir,omitempty" json:"sparseCheckoutDir,omitempty"`

	// The tag or commit id to reset the checked out branch to.
	StartPoint string `yaml:"startPoint,omitempty" json:"startPoint,omitempty"`

	// The name of the tag to reset the checked out branch to.
	//  Note that this is equivalent to 'startPoint' and provided for convenience.
	Tag string `yaml:"tag,omitempty" json:"tag,omitempty"`
}

type Command struct {

	// List of the actions of given command. Now the only one command must be
	// specified in list but there are plans to implement supporting multiple actions
	// commands.
	Actions []CommandAction `yaml:"actions" json:"actions"`

	// Additional command attributes
	Attributes Attributes `yaml:"attributes,omitempty" json:"attributes,omitempty"`

	// Describes the name of the command. Should be unique per commands set.
	Name string `yaml:"name"`

	// Preview url
	PreviewUrl CommandPreviewUrl `yaml:"previewUrl,omitempty" json:"previewUrl,omitempty"`
}

type CommandPreviewUrl struct {
	Port int32  `yaml:"port,omitempty" json:"port,omitempty"`
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
}

type CommandAction struct {

	// The actual action command-line string
	Command string `yaml:"command,omitempty" json:"command,omitempty"`

	// Describes component to which given action relates
	Component string `yaml:"component,omitempty" json:"component,omitempty"`

	// the path relative to the location of the devfile to the configuration file
	// defining one or more actions in the editor-specific format
	Reference string `yaml:"reference,omitempty" json:"reference,omitempty"`

	// The content of the referenced configuration file that defines one or more
	// actions in the editor-specific format
	ReferenceContent string `yaml:"referenceContent,omitempty" json:"referenceContent,omitempty"`

	// Describes action type
	Type CommandType `yaml:"type,omitempty" json:"type,omitempty"`

	// Working directory where the command should be executed
	Workdir string `yaml:"workdir,omitempty" json:"workdir,omitempty"`
}

type Component struct {

	// The name using which other places of this devfile (like commands) can refer to
	// this component. This attribute is optional but must be unique in the devfile if
	// specified.
	Alias string `yaml:"alias,omitempty" json:"alias,omitempty"`

	// Describes whether projects sources should be mount to the component.
	// `CHE_PROJECTS_ROOT` environment variable should contains a path where projects
	// sources are mount
	MountSources bool `yaml:"mountSources,omitempty" json:"mountSources,omitempty"`

	// Describes type of the component, e.g. whether it is an plugin or editor or
	// other type
	Type ComponentType `yaml:"type" json:"type"`

	// for type ChePlugin
	ComponentChePlugin `yaml:",inline" json:",inline"`

	// for type=dockerfile
	ComponentDockerimage `yaml:",inline" json:",inline"`
}

type ComponentChePlugin struct {
	Id          string `yaml:"id,omitempty" json:"id,omitempty"`
	Reference   string `yaml:"reference,omitempty" json:"reference,omitempty"`
	RegistryUrl string `yaml:"registryUrl,omitempty" json:"registryUrl,omitempty"`
}

type ComponentCheEditor struct {
	Id          string `yaml:"id,omitempty" json:"id,omitempty"`
	Reference   string `yaml:"reference,omitempty" json:"reference,omitempty"`
	RegistryUrl string `yaml:"registryUrl,omitempty" json:"registryUrl,omitempty"`
}

type ComponentOpenshift struct {
	Reference        string `yaml:"reference,omitempty" json:"reference,omitempty"`
	ReferenceContent string `yaml:"referenceContent,omitempty" json:"referenceContent,omitempty"`
	Selector         string `yaml:"selector,omitempty" json:"selector,omitempty"`
	EntryPoints      string `yaml:"entryPoints,omitempty" json:"entryPoints,omitempty"`
	MemoryLimit      string `yaml:"memoryLimit,omitempty" json:"memoryLimit,omitempty"`
}

type ComponentKubernetes struct {
	Reference        string `yaml:"reference,omitempty" json:"reference,omitempty"`
	ReferenceContent string `yaml:"referenceContent,omitempty" json:"referenceContent,omitempty"`
	Selector         string `yaml:"selector,omitempty" json:"selector,omitempty"`
	EntryPoints      string `yaml:"entryPoints,omitempty" json:"entryPoints,omitempty"`
	MemoryLimit      string `yaml:"memoryLimit,omitempty" json:"memoryLimit,omitempty"`
}

type ComponentDockerimage struct {
	Image       string                `yaml:"image,omitempty" json:"image,omitempty"`
	MemoryLimit string                `yaml:"memoryLimit,omitempty" json:"memoryLimit,omitempty"`
	Command     []string              `yaml:"command,omitempty" json:"command,omitempty"`
	Args        []string              `yaml:"args,omitempty" json:"args,omitempty"`
	Volumes     []DockerimageVolume   `yaml:"volumes,omitempty" json:"volumes,omitempty"`
	Env         []DockerimageEnv      `yaml:"env,omitempty" json:"env,omitempty"`
	Endpoints   []DockerimageEndpoint `yaml:"endpoints,omitempty" json:"endpoints,omitempty"`
}

type DockerimageVolume struct {
	Name          string `yaml:"name,omitempty" json:"name,omitempty"`
	ContainerPath string `yaml:"containerPath,omitempty" json:"containerPath,omitempty"`
}

type DockerimageEnv struct {
	Name  string `yaml:"name,omitempty" json:"name,omitempty"`
	Value string `yaml:"value,omitempty" json:"value,omitempty"`
}

type DockerimageEndpoint struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	Port int32  `yaml:"port,omitempty" json:"port,omitempty"`
}
