// package asker uses the Survey library to interact with the user and ask various information
// needed to initiate a project.
package asker

import (
	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/registry"
)

// Asker interactively asks for information to the user
type Asker interface {
	// AskArchitectures asks for a selection of architectures from a list of architecture names
	AskArchitectures(archs []string, selectedDefault []string) ([]string, error)

	// AskLanguage asks for a language, from a list of language names.
	// back is returned as true if the user selected to go back, or the language name is returned
	AskLanguage(langs []string) (back bool, result string, err error)

	// AskType asks for a Devfile type, or to go back. back is returned as true if the user selected to go back,
	// or the selected type is returned
	AskType(types registry.TypesWithDetails) (back bool, _ api.DevfileStack, _ error)

	// AskVersion asks for the Devfile version, or to go back. back is returned as true if the user selected to go back,
	// or the selected version is returned
	AskVersion(versions []api.DevfileStackVersion) (back bool, version string, _ error)

	// AskStarterProject asks for an optional project, from a list of projects. If no project is selected, false is returned.
	// Or the index of the selected project is returned
	AskStarterProject(projects []string) (selected bool, _ int, _ error)

	// AskName asks for a devfile component name
	AskName(defaultName string) (string, error)

	// AskCorrect asks for confirmation
	AskCorrect() (bool, error)

	AskContainerName(containers []string) (string, error)

	// AskPersonalizeConfiguration asks the configuration user wants to change
	AskPersonalizeConfiguration(configuration ContainerConfiguration) (OperationOnContainer, error)

	// AskAddEnvVar asks the key and value for env var
	AskAddEnvVar() (string, string, error)

	// AskAddPort asks the container name and port that user wants to add
	AskAddPort() (string, error)
}

type ContainerConfiguration struct {
	Ports []string
	Envs  map[string]string
}

type OperationOnContainer struct {
	Ops  string
	Kind string
	Key  string
}

// key is container name
type DevfileConfiguration map[string]ContainerConfiguration
