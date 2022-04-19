package backend

type CreateBindingBackend interface {
	// Validate returns error if the backend failed to validate; mainly useful for flags backend
	Validate(flags map[string]string) error
	// SelectServiceInstance lists services that can be binded to the component
	SelectServiceInstance(flags map[string]string, options []string) (string, error)
	// AskBindingName asks for the service name to be set
	AskBindingName(componentName string, flags map[string]string) (string, error)
	// AskBindAsFiles asks if service should be binded as files
	AskBindAsFiles(flags map[string]string) (bool, error)
}
