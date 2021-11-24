package common

// ApplyClient is a wrapper around ApplyComponent which runs an apply command on a component
type ApplyClient interface {
	ApplyComponent(component string) error
}
