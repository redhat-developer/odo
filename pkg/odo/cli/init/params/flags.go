package params

import "github.com/redhat-developer/odo/pkg/odo/util"

// FlagsBuilder is a backend that will extract init parameters from flags passed to the command
type FlagsBuilder struct {
	flags map[string]string
}

func (o *FlagsBuilder) IsAdequate(flags map[string]string) bool {
	// Save args to use them for building params
	o.flags = flags
	return len(flags) > 0
}

func (o *FlagsBuilder) ParamsBuild() (InitParams, error) {
	if len(o.flags) == 0 {
		util.LogErrorAndExit(nil, "IsAdequate must be called and return true before to call ParamsBuild")
	}
	return InitParams{
		Name:            o.flags[FLAG_NAME],
		Devfile:         o.flags[FLAG_DEVFILE],
		DevfileRegistry: o.flags[FLAG_DEVFILE_REGISTRY],
		Starter:         o.flags[FLAG_STARTER],
		DevfilePath:     o.flags[FLAG_DEVFILE_PATH],
	}, nil
}
