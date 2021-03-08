package v2

import (
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// GetParent returns the Parent object parsed from devfile
func (d *DevfileV2) GetParent() *v1.Parent {
	return d.Parent
}

// SetParent sets the parent for the devfile
func (d *DevfileV2) SetParent(parent *v1.Parent) {
	d.Parent = parent
}
