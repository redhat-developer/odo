package service

import (
	"fmt"
	"path/filepath"

	svc "github.com/openshift/odo/pkg/service"

	"github.com/openshift/odo/pkg/odo/cli/component"
	"github.com/openshift/odo/pkg/util"
)

// validDevfileDirectory returns an error if the "odo service" command is executed from a directory not containing devfile.yaml
func validDevfileDirectory(componentContext string) error {
	if componentContext == "" {
		componentContext = component.LocalDirectoryDefaultLocation
	}
	devfilePath := filepath.Join(componentContext, component.DevfilePath)
	if !util.CheckPathExists(devfilePath) {
		return fmt.Errorf("service can be created/deleted from a valid component directory only\n"+
			"refer %q for more information", "odo servce create -h")
	}
	return nil
}

// decideBackend returns the type of service provider backend to be used
func decideBackend(arg string) ServiceProviderBackend {
	_, _, err := svc.SplitServiceKindName(arg)
	if err != nil {
		// failure to split provided name into two; hence ServiceCatalogBackend
		return NewServiceCatalogBackend()
	} else {
		// provided name adheres to the format <operator-type>/<crd-name>; hence OperatorBackend
		return NewOperatorBackend()
	}
}
