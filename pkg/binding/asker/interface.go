package asker

type CreationOption int

const (
	CreateOnCluster CreationOption = iota
	OutputToStdout
	OutputToFile
)

type ServiceInstancesNamespaceListOption int

const (
	CurrentNamespace ServiceInstancesNamespaceListOption = iota + 1
	AllAccessibleNamespaces
)

type Asker interface {
	// SelectNamespaceListOption asks the user to select how service instances should be listed,
	// i.e. from the current namespace or from all accessible namespaces.
	// If user selects to list from all accessible namespaces, callers should fetch the list of namespaces
	// and prompt users to either pick or enter one.
	// See SelectNamespace and AskNamespace methods.
	SelectNamespaceListOption() (ServiceInstancesNamespaceListOption, error)
	// AskNamespace asks the user to enter a namespace
	AskNamespace() (string, error)
	// SelectNamespace takes a list of namespaces and asks the user to pick one of them
	SelectNamespace(options []string) (string, error)
	// SelectWorkloadResource takes a list of workloads resources and asks user to select one
	SelectWorkloadResource(options []string) (int, error)
	// SelectWorkloadResourceName takes a list of workloads resources names and asks user to select one
	// returns back true if the user selected to go back
	SelectWorkloadResourceName(names []string) (back bool, selected string, err error)
	// SelectNamingStrategy asks for the naming strategy to be used from a list of options.
	// If NamingStrategyCustom is returned, it means the user needs to be prompted for a value. See AskNamingStrategy.
	SelectNamingStrategy() (string, error)
	// AskWorkloadResourceName asks user to type resource name
	AskWorkloadResourceName() (string, error)
	// AskServiceInstance takes a list of services and asks user to select one
	AskServiceInstance(serviceInstanceOptions []string) (string, error)
	// AskServiceBindingName asks for service binding name to be set
	AskServiceBindingName(defaultName string) (string, error)
	// AskBindAsFiles asks if service should be binded as files
	AskBindAsFiles() (bool, error)
	// AskNamingStrategy prompts for the naming strategy to be used
	AskNamingStrategy() (string, error)
	// SelectCreationOptions asks to select creation options for the servicebinding
	SelectCreationOptions() ([]CreationOption, error)
	// AskOutputFilePath asks for the path of the file to output service binding
	AskOutputFilePath(defaultValue string) (string, error)
}
