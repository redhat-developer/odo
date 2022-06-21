package asker

type CreationOption int

const (
	CreateOnCluster CreationOption = iota
	OutputToStdout
	OutputToFile
)

type Asker interface {
	// SelectWorkloadResource takes a list of workloads resources and asks user to select one
	SelectWorkloadResource(options []string) (int, error)
	// SelectWorkloadResourceName takes a list of workloads resources names and asks user to select one
	SelectWorkloadResourceName(names []string) (string, error)
	// AskWorkloadResourceName asks user to type resource name
	AskWorkloadResourceName() (string, error)
	// AskServiceInstance takes a list of services and asks user to select one
	AskServiceInstance(serviceInstanceOptions []string) (string, error)
	// AskServiceBindingName asks for service binding name to be set
	AskServiceBindingName(defaultName string) (string, error)
	// AskBindAsFiles asks if service should be binded as files
	AskBindAsFiles() (bool, error)
	// SelectCreationOption asks to select a creation option for the servicebinding
	SelectCreationOption() ([]CreationOption, error)
	// AskOutputFilePath asks for the path of the file to output service binding
	AskOutputFilePath(defaultValue string) (string, error)
}
