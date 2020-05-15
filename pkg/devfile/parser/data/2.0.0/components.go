package version200

import (
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// GetComponents returns the slice of DevfileComponent objects parsed from the Devfile
func (d *Devfile200) GetComponents() []common.DevfileComponent {
	return d.Components
}

// GetCommands returns the slice of DevfileCommand objects parsed from the Devfile
func (d *Devfile200) GetCommands() []common.DevfileCommand {
	return d.Commands
}

// GetParent returns the  DevfileParent object parsed from devfile
func (d *Devfile200) GetParent() common.DevfileParent {
	return d.Parent
}

// GetProject returns the DevfileProject Object parsed from devfile
func (d *Devfile200) GetProject() []common.DevfileProject {
	return d.Projects
}

// GetMetadata returns the DevfileMetadata Object parsed from devfile
func (d *Devfile200) GetMetadata() common.DevfileMetadata {
	return d.Metadata
}

// GetEvents returns the Events Object parsed from devfile
func (d *Devfile200) GetEvents() common.DevfileEvents {
	return d.Events
}

// GetAliasedComponents returns the slice of DevfileComponent objects that each have an alias
func (d *Devfile200) GetAliasedComponents() []common.DevfileComponent {
	// V2 has name required in jsonSchema
	return d.Components
}
