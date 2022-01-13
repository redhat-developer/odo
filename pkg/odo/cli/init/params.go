package init

import "errors"

type initParams struct {
	// name of the component to create (required)
	name string
	// name of the devfile in devfile registry (required if --devfile-path is not defined)
	devfile string
	// name of the devfile registry (as configured in odo registry). It can be used in combination with --devfile, but not with --devfile-path (optional)
	devfileRegistry string
	// name of the starter project (optional)
	starter string
	// path to a devfile. This is alternative to using devfile from Devfile registry. It can be local filesystem path or http(s) URL (required if --devfile is not defined)
	devfilePath string
}

func (o *initParams) validate() error {
	if o.name == "" {
		return errors.New("name is required")
	}
	if o.devfile == "" && o.devfilePath == "" {
		return errors.New("Either devfile or devfile-path should be set")
	}
	if o.devfile != "" && o.devfilePath != "" {
		return errors.New("Only one of devfile or devfile-path should be set")
	}
	if o.devfilePath != "" && o.devfileRegistry != "" {
		return errors.New("devfile-registry cannot be used with devfile-path")
	}
	return nil
}

type ParamsBuilder interface {
	// IsAdequate returns true if the implementation is able to build parameters, given the arguments passed to the comamnd
	IsAdequate(flags map[string]string) bool
	// ParamsBuild returns parameters for init
	ParamsBuild() (initParams, error)
}
