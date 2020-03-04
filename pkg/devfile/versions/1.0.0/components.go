package version100

import (
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
