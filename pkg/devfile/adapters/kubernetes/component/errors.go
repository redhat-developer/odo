package component

type NoDefaultDeployCommandFoundError struct{}

func (e NoDefaultDeployCommandFoundError) Error() string {
	return "error deploying, no default deploy command found in devfile"
}

type MoreThanOneDefaultDeployCommandFoundError struct{}

func (e MoreThanOneDefaultDeployCommandFoundError) Error() string {
	return "more than one default deploy command found in devfile, should not happen"
}
