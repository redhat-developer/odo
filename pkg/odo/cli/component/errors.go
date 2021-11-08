package component

import "fmt"

type DevfileExistsDevfileFlagError struct{}

func (e *DevfileExistsDevfileFlagError) Error() string {
	return "this directory already contains a devfile, you can't specify devfile via --devfile"
}

type DevfileExistsExtraArgsError struct {
	args int
}

func (e *DevfileExistsExtraArgsError) Error() string {
	return fmt.Sprintf("accepts between 0 and 1 arg when using existing devfile, received %d", e.args)
}

type DevfileFlagWithRegistryFlagError struct{}

func (e *DevfileFlagWithRegistryFlagError) Error() string {
	return "you can't specify registry via --registry if you want to use the devfile that is specified via --devfile"
}
