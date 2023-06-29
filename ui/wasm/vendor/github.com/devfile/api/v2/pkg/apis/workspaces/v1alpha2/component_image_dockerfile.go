package v1alpha2

// DockerfileSrcType describes the type of
// the src for the Dockerfile outerloop build.
// Only one of the following location type may be specified.
// +kubebuilder:validation:Enum=Uri;DevfileRegistry;Git
type DockerfileSrcType string

const (
	UriLikeDockerfileSrcType             DockerfileSrcType = "Uri"
	DevfileRegistryLikeDockerfileSrcType DockerfileSrcType = "DevfileRegistry"
	GitLikeDockerfileSrcType             DockerfileSrcType = "Git"
)

// Dockerfile Image type to specify the outerloop build using a Dockerfile
type DockerfileImage struct {
	BaseImage     `json:",inline"`
	DockerfileSrc `json:",inline"`
	Dockerfile    `json:",inline"`
}

// +union
type DockerfileSrc struct {
	// Type of Dockerfile src
	// +
	// +unionDiscriminator
	// +optional
	SrcType DockerfileSrcType `json:"srcType,omitempty"`

	// URI Reference of a Dockerfile.
	// It can be a full URL or a relative URI from the current devfile as the base URI.
	// +optional
	Uri string `json:"uri,omitempty"`

	// Dockerfile's Devfile Registry source
	// +optional
	DevfileRegistry *DockerfileDevfileRegistrySource `json:"devfileRegistry,omitempty"`

	// Dockerfile's Git source
	// +optional
	Git *DockerfileGitProjectSource `json:"git,omitempty"`
}

// +devfile:getter:generate
type Dockerfile struct {
	// Path of source directory to establish build context. Defaults to ${PROJECT_SOURCE} in the container
	// +optional
	BuildContext string `json:"buildContext,omitempty"`

	// The arguments to supply to the dockerfile build.
	// +optional
	Args []string `json:"args,omitempty" patchStrategy:"replace"`

	// Specify if a privileged builder pod is required.
	//
	// Default value is `false`
	// +optional
	// +devfile:default:value=false
	RootRequired *bool `json:"rootRequired,omitempty"`
}

type DockerfileDevfileRegistrySource struct {
	// Id in a devfile registry that contains a Dockerfile. The src in the OCI registry
	// required for the Dockerfile build will be downloaded for building the image.
	Id string `json:"id"`

	// Devfile Registry URL to pull the Dockerfile from when using the Devfile Registry as Dockerfile src.
	// To ensure the Dockerfile gets resolved consistently in different environments,
	// it is recommended to always specify the `devfileRegistryUrl` when `Id` is used.
	// +optional
	RegistryUrl string `json:"registryUrl,omitempty"`
}

type DockerfileGitProjectSource struct {
	// Git src for the Dockerfile build. The src required for the Dockerfile build will need to be
	// cloned for building the image.
	GitProjectSource `json:",inline"`

	// Location of the Dockerfile in the Git repository when using git as Dockerfile src.
	// Defaults to Dockerfile.
	// +optional
	FileLocation string `json:"fileLocation,omitempty"`
}
