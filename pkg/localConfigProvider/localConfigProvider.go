package localConfigProvider

// LocalStorage holds storage related information
type LocalStorage struct {
	// Name of the storage
	Name string `yaml:"Name,omitempty"`
	// Size of the storage
	Size string `yaml:"Size,omitempty"`
	// Boolean indicating if the volume should be ephemeral. A nil pointer indicates to use the default behaviour
	Ephemeral *bool `yaml:"Ephemeral,omitempty"`
	// Path of the storage to which it will be mounted on the container
	Path string `yaml:"Path,omitempty"`
	// Container is the container name on which this storage is mounted
	Container string `yaml:"-" json:"-"`
}
