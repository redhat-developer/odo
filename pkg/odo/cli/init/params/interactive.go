package params

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/odo/cli/init/asker"
)

// InteractiveBuilder is a backend that will ask init parameters interactively
type InteractiveBuilder struct {
	asker         asker.Asker
	catalogClient catalog.Client
}

func NewInteractiveBuilder(asker asker.Asker, catalogClient catalog.Client) *InteractiveBuilder {
	return &InteractiveBuilder{
		asker:         asker,
		catalogClient: catalogClient,
	}
}

func (o *InteractiveBuilder) IsAdequate(flags map[string]string) bool {
	return len(flags) == 0
}

func (o *InteractiveBuilder) ParamsBuild() (InitParams, error) {
	result := InitParams{}
	devfileEntries, _ := o.catalogClient.ListDevfileComponents("")
	langs := devfileEntries.GetLanguages()
	lang, err := o.asker.AskLanguage(langs)
	if err != nil {
		return InitParams{}, err
	}
	types := devfileEntries.GetProjectTypes(lang)
	details, err := o.asker.AskType(types)
	if err != nil {
		return InitParams{}, err
	}
	result.DevfileRegistry = details.Registry.Name
	result.Devfile = details.Name

	projects, err := o.catalogClient.GetStarterProjectsNames(details)
	if err != nil {
		return InitParams{}, err
	}

	result.Starter, err = o.asker.AskStarterProject(projects)
	if err != nil {
		return InitParams{}, err
	}

	result.Name, err = o.asker.AskName(fmt.Sprintf("my-%s-app", result.Devfile))
	if err != nil {
		return InitParams{}, err
	}

	return result, nil
}
