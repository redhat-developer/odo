package genericclioptions

var _ error = NoDevfileError{}

type NoDevfileError struct{}

func (o NoDevfileError) Error() string {
	return "no devfile found"
}

func IsNoDevfileError(err error) bool {
	_, ok := err.(NoDevfileError)
	return ok
}
