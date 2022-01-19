package params

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/odo/cli/init/asker"
)

const (
	STATE_ASK_LANG = iota
	STATE_ASK_TYPE
	STATE_ASK_STARTER
	STATE_END
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
	state := STATE_ASK_LANG
	var lang string
	var err error
	var details catalog.DevfileComponentType
loop:
	for {
		switch state {

		case STATE_ASK_LANG:
			lang, err = o.asker.AskLanguage(langs)
			if err != nil {
				return InitParams{}, err
			}
			state = STATE_ASK_TYPE

		case STATE_ASK_TYPE:
			types := devfileEntries.GetProjectTypes(lang)
			var back bool
			back, details, err = o.asker.AskType(types)
			if err != nil {
				return InitParams{}, err
			}
			if back {
				state = STATE_ASK_LANG
				continue loop
			}
			result.DevfileRegistry = details.Registry.Name
			result.Devfile = details.Name
			state = STATE_ASK_STARTER

		case STATE_ASK_STARTER:
			projects, err := o.catalogClient.GetStarterProjectsNames(details)
			if err != nil {
				return InitParams{}, err
			}
			var back bool
			back, result.Starter, err = o.asker.AskStarterProject(projects)
			if err != nil {
				return InitParams{}, err
			}
			if back {
				state = STATE_ASK_TYPE
				continue loop
			}

			result.Name, err = o.asker.AskName(fmt.Sprintf("my-%s-app", result.Devfile))
			if err != nil {
				return InitParams{}, err
			}
			state = STATE_END

		case STATE_END:
			break loop
		}
	}

	return result, nil
}
