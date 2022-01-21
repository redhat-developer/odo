package params

import (
	"errors"
	"fmt"

	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/util"
)

type InitParams struct {
	// Name of the component to create (required)
	Name string
	// name of the Devfile in Devfile registry (required if --Devfile-path is not defined)
	Devfile string
	// name of the devfile registry (as configured in odo registry). It can be used in combination with --devfile, but not with --devfile-path (optional)
	DevfileRegistry string
	// name of the Starter project (optional)
	Starter string
	// path to a devfile. This is alternative to using devfile from Devfile registry. It can be local filesystem path or http(s) URL (required if --devfile is not defined)
	DevfilePath string
}

func (o *InitParams) Validate(prefClient preference.Client) error {
	if o.Name == "" {
		return errors.New("name is required")
	}
	if o.Devfile == "" && o.DevfilePath == "" {
		return errors.New("either devfile or devfile-path should be set")
	}
	if o.Devfile != "" && o.DevfilePath != "" {
		return errors.New("only one of devfile or devfile-path should be set")
	}

	if o.DevfileRegistry != "" && !prefClient.RegistryNameExists(o.DevfileRegistry) {
		return fmt.Errorf("registry %q not found in the list of devfile registries. Please use `odo registry` command to configure devfile registries", o.DevfileRegistry)
	}

	if o.DevfilePath != "" && o.DevfileRegistry != "" {
		return errors.New("devfile-registry cannot be used with devfile-path")
	}

	err := util.ValidateK8sResourceName("name", o.Name)
	if err != nil {
		return err
	}

	return nil
}
