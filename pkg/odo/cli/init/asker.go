package init

import "github.com/redhat-developer/odo/pkg/catalog"

type asker interface {
	askLanguage(langs []string) (string, error)
	askType(types catalog.TypesWithDetails) (catalog.DevfileComponentType, error)
	askStarterProject(projects []string) (string, error)
	askName(defaultName string) (string, error)
}
