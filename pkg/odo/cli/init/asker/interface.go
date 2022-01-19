package asker

import "github.com/redhat-developer/odo/pkg/catalog"

type Asker interface {
	AskLanguage(langs []string) (string, error)
	AskType(types catalog.TypesWithDetails) (back bool, _ catalog.DevfileComponentType, _ error)
	AskStarterProject(projects []string) (back bool, _ string, _ error)
	AskName(defaultName string) (string, error)
}
