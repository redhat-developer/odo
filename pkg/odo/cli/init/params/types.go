package params

import "errors"

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

func (o *InitParams) Validate() error {
	if o.Name == "" {
		return errors.New("name is required")
	}
	if o.Devfile == "" && o.DevfilePath == "" {
		return errors.New("Either devfile or devfile-path should be set")
	}
	if o.Devfile != "" && o.DevfilePath != "" {
		return errors.New("Only one of devfile or devfile-path should be set")
	}
	if o.DevfilePath != "" && o.DevfileRegistry != "" {
		return errors.New("devfile-registry cannot be used with devfile-path")
	}
	return nil
}
