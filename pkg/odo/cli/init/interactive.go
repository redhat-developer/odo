package init

// InteractiveBuilder is a backend that will ask init parameters interactively
type InteractiveBuilder struct{}

func (o *InteractiveBuilder) IsAdequate(flags map[string]string) bool {
	return len(flags) == 0
}

func (o *InteractiveBuilder) ParamsBuild() (initParams, error) {
	return initParams{}, nil
}
