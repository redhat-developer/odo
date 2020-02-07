package version100

import (
	"github.com/openshift/odo/pkg/devfile/versions/common"
)

// GetComponents returns a slice of DevfileComponent objects
func (d *Devfile100) GetComponents() []common.DevfileComponent {
	return d.Components
}
