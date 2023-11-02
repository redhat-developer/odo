package v1alpha2

// Component that allows the developer to declare and configure a volume into their devworkspace
type VolumeComponent struct {
	BaseComponent `json:",inline"`
	Volume        `json:",inline"`
}

// Volume that should be mounted to a component container
// +devfile:getter:generate
type Volume struct {
	// +optional
	// Size of the volume
	Size string `json:"size,omitempty"`

	// +optional
	// Ephemeral volumes are not stored persistently across restarts. Defaults
	// to false
	// +devfile:default:value=false
	Ephemeral *bool `json:"ephemeral,omitempty"`
}
