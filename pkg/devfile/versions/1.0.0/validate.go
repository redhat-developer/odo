package version100

import (
	"github.com/golang/glog"
	"github.com/openshift/odo/pkg/devfile/versions/common"
)

// Validate validates whether sections of devfile are odo compatible
func (d *Devfile100) Validate() error {

	// Validate Projects
	if err := common.ValidateProjects(d.Projects); err != nil {
		return err
	}

	// Validate Components
	if err := common.ValidateComponents(d.Components); err != nil {
		return err
	}

	// Validate Commands
	if err := common.ValidateCommands(d.Commands); err != nil {
		return err
	}

	// Successful
	glog.V(4).Info("Successfully validated devfile sections for apiVersion 1.0.0")
	return nil
}
