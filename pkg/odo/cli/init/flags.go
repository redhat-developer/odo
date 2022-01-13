package init

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

func (o *FlagsBuilder) ParamsBuild() (initParams, error) {
	if len(o.flags) == 0 {
		util.LogErrorAndExit(nil, "IsAdequate must be called and return true before to call ParamsBuild")
	}
	return initParams{
		name:            o.flags[FLAG_NAME],
		devfile:         o.flags[FLAG_DEVFILE],
		devfileRegistry: o.flags[FLAG_DEVFILE_REGISTRY],
		starter:         o.flags[FLAG_STARTER],
		devfilePath:     o.flags[FLAG_DEVFILE_PATH],
	}, nil
}
