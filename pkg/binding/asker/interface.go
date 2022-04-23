package asker

type Asker interface {
	// AskServiceInstance takes a list of services and asks user to select one
	AskServiceInstance(serviceInstanceOptions []string) (string, error)
	// AskServiceBindingName asks for service binding name to be set
	AskServiceBindingName(defaultName string) (string, error)
	// AskBindAsFiles asks if service should be binded as files
	AskBindAsFiles() (bool, error)
}
