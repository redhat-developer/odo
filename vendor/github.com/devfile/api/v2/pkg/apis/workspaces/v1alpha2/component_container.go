package v1alpha2

// Component that allows the developer to add a configured container into their devworkspace
type ContainerComponent struct {
	BaseComponent `json:",inline"`
	Container     `json:",inline"`
	Endpoints     []Endpoint `json:"endpoints,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

// Annotation specifies the annotations to be added to specific resources
type Annotation struct {
	// +optional
	// Annotations to be added to deployment
	Deployment map[string]string `json:"deployment,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// +optional
	// Annotations to be added to service
	Service map[string]string `json:"service,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

// +devfile:getter:generate
type Container struct {
	Image string `json:"image"`

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
	Env []EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// +optional
	// Annotations that should be added to specific resources for this container
	Annotation *Annotation `json:"annotation,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// +optional
	// List of volumes mounts that should be mounted is this container.
	VolumeMounts []VolumeMount `json:"volumeMounts,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

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
	// +kubebuilder:default=/projects
	SourceMapping string `json:"sourceMapping,omitempty"`

	// Specify if a container should run in its own separated pod,
	// instead of running as part of the main development environment pod.
	//
	// Default value is `false`
	// +optional
	// +devfile:default:value=false
	DedicatedPod *bool `json:"dedicatedPod,omitempty"`
}

//GetMountSources returns the value of the boolean property.  If it's unset, the default value is true for all component types except plugins and components that set `dedicatedPod` to true.
func (in *Container) GetMountSources() bool {
	if in.MountSources != nil {
		return *in.MountSources
	} else {
		if in.GetDedicatedPod() {
			return false
		}
		return true
	}
}

type EnvVar struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

// Volume that should be mounted to a component container
type VolumeMount struct {
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
