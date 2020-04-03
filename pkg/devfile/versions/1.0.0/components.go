package version100

import (
	"strings"

	"github.com/openshift/odo/pkg/devfile/versions/common"
)

// GetComponents returns the slice of DevfileComponent objects parsed from the Devfile
func (d *Devfile100) GetComponents() []common.DevfileComponent {
	return d.Components
}

// GetAliasedComponents returns the slice of DevfileComponent objects that each have an alias
func (d *Devfile100) GetAliasedComponents() []common.DevfileComponent {
	var aliasedComponents = []common.DevfileComponent{}
	for _, comp := range d.Components {
		if comp.Alias != nil {
			aliasedComponents = append(aliasedComponents, comp)
		}
	}
	return aliasedComponents
}

// GetProjects returns the slice of DevfileProject objects parsed from the Devfile
func (d *Devfile100) GetProjects() []common.DevfileProject {
	return d.Projects
}

// GetCommands returns the slice of DevfileCommand objects parsed from the Devfile
func (d *Devfile100) GetCommands() []common.DevfileCommand {
	var commands []common.DevfileCommand

	for _, command := range d.Commands {
		command.Name = strings.ToLower(command.Name)
		commands = append(commands, command)
	}

	return commands
}
