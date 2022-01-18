package params

const FLAG_NAME = "name"
const FLAG_DEVFILE = "devfile"
const FLAG_DEVFILE_REGISTRY = "devfile-registry"
const FLAG_STARTER = "starter"
const FLAG_DEVFILE_PATH = "devfile-path"

type ParamsBuilder interface {
	// IsAdequate returns true if the implementation is able to build parameters, given the arguments passed to the comamnd
	IsAdequate(flags map[string]string) bool
	// ParamsBuild returns parameters for init
	ParamsBuild() (InitParams, error)
}
