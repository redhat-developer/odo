package init

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/catalog"
)

// InteractiveBuilder is a backend that will ask init parameters interactively
type InteractiveBuilder struct {
	asker         asker
	catalogClient catalog.Client
}

func NewInteractiveBuilder(asker asker, catalogClient catalog.Client) *InteractiveBuilder {
	return &InteractiveBuilder{
		asker:         asker,
		catalogClient: catalogClient,
	}
}

func (o *InteractiveBuilder) IsAdequate(flags map[string]string) bool {
	return len(flags) == 0
}

func (o *InteractiveBuilder) ParamsBuild() (initParams, error) {
	result := initParams{}
	devfileEntries, _ := o.catalogClient.ListDevfileComponents("")
	langs := devfileEntries.GetLanguages()
	lang, err := o.asker.askLanguage(langs)
	if err != nil {
		return initParams{}, err
	}
	types := devfileEntries.GetProjectTypes(lang)
	details, err := o.asker.askType(types)
	if err != nil {
		return initParams{}, err
	}
	result.devfileRegistry = details.Registry.Name
	result.devfile = details.Name

	projects, err := o.catalogClient.GetStarterProjectsNames(details)
	if err != nil {
		return initParams{}, err
	}

	result.starter, err = o.asker.askStarterProject(projects)
	if err != nil {
		return initParams{}, err
	}

	result.name, err = o.asker.askName(fmt.Sprintf("my-%s-app", result.devfile))
	if err != nil {
		return initParams{}, err
	}

	return result, nil
}
