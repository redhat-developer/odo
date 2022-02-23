package backend

import (
	"fmt"
	"reflect"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/alizer/go/pkg/apis/recognizer"
	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/init/asker"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

type AlizerBackend struct {
	askerClient   asker.Asker
	catalogClient catalog.Client
}

func NewAlizerBackend(askerClient asker.Asker, catalogClient catalog.Client) *AlizerBackend {
	return &AlizerBackend{
		askerClient:   askerClient,
		catalogClient: catalogClient,
	}
}

func (o *AlizerBackend) Validate(flags map[string]string, fs filesystem.Filesystem, dir string) error {
	return nil
}

// detectFramework uses the alizer library in order to detect the devfile
// to use depending on the files in the path
func (o *AlizerBackend) detectFramework(path string) (recognizer.DevFileType, catalog.Registry, error) {
	types := []recognizer.DevFileType{}
	components, err := o.catalogClient.ListDevfileComponents("")
	if err != nil {
		return recognizer.DevFileType{}, catalog.Registry{}, err
	}
	for _, component := range components.Items {
		types = append(types, recognizer.DevFileType{
			Name:        component.Name,
			Language:    component.Language,
			ProjectType: component.ProjectType,
			Tags:        component.Tags,
		})
	}
	typ, err := recognizer.SelectDevFileFromTypes(path, types)
	if err != nil {
		return recognizer.DevFileType{}, catalog.Registry{}, err
	}

	// TODO(feloy): This part won't be necessary when SelectDevFileFromTypes returns the index
	var indexOfDetected int
	for i, typeFromList := range types {
		if reflect.DeepEqual(typeFromList, typ) {
			indexOfDetected = i
			break
		}
	}
	registry := components.Items[indexOfDetected].Registry
	return typ, registry, nil
}

// SelectDevfile calls thz Alizer to detect the devfile and asks for confirmation to the user
func (o *AlizerBackend) SelectDevfile(flags map[string]string, fs filesystem.Filesystem, dir string) (location *DevfileLocation, err error) {
	selected, registry, err := o.detectFramework(dir)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Based on the files in the current directory odo detected\nLanguage: %s\nProject type: %s\n", selected.Language, selected.ProjectType)
	fmt.Printf("The devfile %q from the registry %q will be downloaded.\n", selected.Name, registry.Name)
	confirm, err := o.askerClient.AskCorrect()
	if err != nil {
		return nil, err
	}
	if !confirm {
		return nil, nil
	}
	return &DevfileLocation{
		Devfile:         selected.Name,
		DevfileRegistry: registry.Name,
	}, nil
}

func (o *AlizerBackend) SelectStarterProject(devfile parser.DevfileObj, flags map[string]string) (starter *v1alpha2.StarterProject, err error) {
	return nil, nil
}

func (o *AlizerBackend) PersonalizeName(devfile parser.DevfileObj, flags map[string]string) error {
	return nil
}
