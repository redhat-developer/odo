package backend

import (
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"

	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/init/asker"
)

const (
	STATE_ASK_LANG = iota
	STATE_ASK_TYPE
	STATE_END
)

// InteractiveBackend is a backend that will ask information interactively using the `asker` package
type InteractiveBackend struct {
	asker         asker.Asker
	catalogClient catalog.Client
}

func NewInteractiveBackend(asker asker.Asker, catalogClient catalog.Client) *InteractiveBackend {
	return &InteractiveBackend{
		asker:         asker,
		catalogClient: catalogClient,
	}
}

func (o *InteractiveBackend) Validate(flags map[string]string) error {
	return nil
}

func (o *InteractiveBackend) SelectDevfile(flags map[string]string) (*DevfileLocation, error) {
	result := &DevfileLocation{}
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
				return nil, err
			}
			state = STATE_ASK_TYPE

		case STATE_ASK_TYPE:
			types := devfileEntries.GetProjectTypes(lang)
			var back bool
			back, details, err = o.asker.AskType(types)
			if err != nil {
				return nil, err
			}
			if back {
				state = STATE_ASK_LANG
				continue loop
			}
			result.DevfileRegistry = details.Registry.Name
			result.Devfile = details.Name
			state = STATE_END
		case STATE_END:
			break loop
		}
	}

	return result, nil
}

func (o *InteractiveBackend) SelectStarterProject(devfile parser.DevfileObj, flags map[string]string) (*v1alpha2.StarterProject, error) {
	starterProjects, err := devfile.Data.GetStarterProjects(parsercommon.DevfileOptions{})
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(starterProjects))
	for _, starterProject := range starterProjects {
		names = append(names, starterProject.Name)
	}

	ok, starter, err := o.asker.AskStarterProject(names)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return &starterProjects[starter], nil
}

func (o *InteractiveBackend) PersonalizeName(devfile parser.DevfileObj, flags map[string]string) error {
	name, err := o.asker.AskName(fmt.Sprintf("my-%s-app", devfile.Data.GetMetadata().Name))
	if err != nil {
		return err
	}
	return devfile.SetMetadataName(name)
}
