package asker

import "github.com/redhat-developer/odo/pkg/catalog"

type Asker interface {
	AskLanguage(langs []string) (string, error)
	AskType(types catalog.TypesWithDetails) (catalog.DevfileComponentType, error)
	AskStarterProject(projects []string) (string, error)
	AskName(defaultName string) (string, error)
}
