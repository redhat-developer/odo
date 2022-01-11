package component

type DefaultProjectError struct{}

func (e *DefaultProjectError) Error() string {
	return "odo may not work as expected in the default project, please run the odo component in a non-default project"
}
