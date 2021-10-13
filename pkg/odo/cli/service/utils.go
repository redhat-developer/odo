package service

import (
	"fmt"

	"github.com/openshift/odo/pkg/devfile/location"
	"github.com/openshift/odo/pkg/odo/cli/component"
	"github.com/openshift/odo/pkg/util"
)

// validDevfileDirectory returns an error if the "odo service" command is executed from a directory not containing devfile.yaml
func validDevfileDirectory(componentContext string) error {
	if componentContext == "" {
		componentContext = component.LocalDirectoryDefaultLocation
	}
	devfilePath := location.DevfileLocation(componentContext)
	if !util.CheckPathExists(devfilePath) {
		return fmt.Errorf("service can be created/deleted from a valid component directory only\n"+
			"refer %q for more information", "odo service create -h")
	}
	return nil
}
