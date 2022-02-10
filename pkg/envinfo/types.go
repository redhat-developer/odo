package envinfo

import (
	"github.com/redhat-developer/odo/pkg/localConfigProvider"
)

// ComponentSettings holds all component related information
type ComponentSettings struct {
	Name string `yaml:"Name,omitempty" json:"name,omitempty"`

	Project string `yaml:"Project,omitempty" json:"project,omitempty"`

	UserCreatedDevfile bool `yaml:"UserCreatedDevfile,omitempty" json:"UserCreatedDevfile,omitempty"`

	URL *[]localConfigProvider.LocalURL `yaml:"Url,omitempty" json:"url,omitempty"`
	// AppName is the application name. Application is a virtual concept present in odo used
	// for grouping of components. A namespace can contain multiple applications
	AppName string `yaml:"AppName,omitempty" json:"appName,omitempty"`

	// DebugPort controls the port used by the pod to run the debugging agent on
	DebugPort *int `yaml:"DebugPort,omitempty" json:"debugPort,omitempty"`

	// RunMode indicates the mode of run used for a successful push
	RunMode *RUNMode `yaml:"RunMode,omitempty" json:"runMode,omitempty"`
}
