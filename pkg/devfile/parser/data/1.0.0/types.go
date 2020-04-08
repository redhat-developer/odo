package version100

import (
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// Devfile100 struct maps to devfile 1.0.0 version schema
type Devfile100 struct {

	// Devfile section "apiVersion"
	ApiVersion common.ApiVersion `yaml:"apiVersion" json:"apiVersion"`

	// Devfile section "metadata"
	Metadata common.DevfileMetadata `yaml:"metadata" json:"metadata"`

	// Devfile section projects
	Projects []common.DevfileProject `yaml:"projects,omitempty" json:"projects,omitempty"`

	Attributes common.Attributes `yaml:"attributes,omitempty" json:"attributes,omitempty"`

	// Description of the workspace components, such as editor and plugins
	Components []common.DevfileComponent `yaml:"components,omitempty" json:"components,omitempty"`

	// Description of the predefined commands to be available in workspace
	Commands []common.DevfileCommand `yaml:"commands,omitempty" json:"commands,omitempty"`
}
