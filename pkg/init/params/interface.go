package params

const FLAG_NAME = "name"
const FLAG_DEVFILE = "devfile"
const FLAG_DEVFILE_REGISTRY = "devfile-registry"
const FLAG_STARTER = "starter"
const FLAG_DEVFILE_PATH = "devfile-path"

// ParamsBuilder builds parameters for the init command, based on various input (either from CLI flags or interactively from user)
type ParamsBuilder interface {
	// IsAdequate returns true if the implementation is able to build parameters, given the arguments passed to the command
	IsAdequate(flags map[string]string) bool
	// ParamsBuild returns parameters for init
	ParamsBuild() (*DevfileLocation, error)
}
