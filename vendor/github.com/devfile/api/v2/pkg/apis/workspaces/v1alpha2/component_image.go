package v1alpha2

// ImageType describes the type of image.
// Only one of the following image type may be specified.
// +kubebuilder:validation:Enum=Dockerfile
type ImageType string

const (
	DockerfileImageType ImageType = "Dockerfile"
)

type BaseImage struct {
}

// Component that allows the developer to build a runtime image for outerloop
type ImageComponent struct {
	BaseComponent `json:",inline"`
	Image         `json:",inline"`
}

type Image struct {
	// Name of the image for the resulting outerloop build
	ImageName  string `json:"imageName"`
	ImageUnion `json:",inline"`
}

// +union
// +devfile:getter:generate
type ImageUnion struct {
	// Type of image
	//
	// +unionDiscriminator
	// +optional
	ImageType ImageType `json:"imageType,omitempty"`

	// Allows specifying dockerfile type build
	// +optional
	Dockerfile *DockerfileImage `json:"dockerfile,omitempty"`

	// Defines if the image should be built during startup.
	//
	// Default value is `false`
	// +optional
	// +devfile:default:value=false
	AutoBuild *bool `json:"autoBuild,omitempty"`
}
