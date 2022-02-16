// package asker uses the Survey library to interact with the user and ask various information
// needed to initiate a project.
package asker

import "github.com/redhat-developer/odo/pkg/catalog"

// Asker interactively asks for information to the user
type Asker interface {
	// AskLanguage asks for a language, from a list of language names. The language name is returned
	AskLanguage(langs []string) (string, error)

	// AskType asks for a Devfile type, or to go back. back is returned as true if the user selected to go back,
	// or the selected type is returned
	AskType(types catalog.TypesWithDetails) (back bool, _ catalog.DevfileComponentType, _ error)

	// AskStarterProject asks for an optional project, from a list of projects. If no project is selected, false is returned.
	// Or the index of the selected project is returned
	AskStarterProject(projects []string) (selected bool, _ int, _ error)

	// AskName asks for a devfile component name
	AskName(defaultName string) (string, error)

	// AskCorrect asks for confirmation
	AskCorrect() (bool, error)
}
