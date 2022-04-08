package alizer

import (
	"reflect"

	"github.com/redhat-developer/alizer/go/pkg/apis/recognizer"
	"github.com/redhat-developer/odo/pkg/registry"
)

type Alizer struct {
	registryClient registry.Client
}

func NewAlizerClient(registryClient registry.Client) *Alizer {
	return &Alizer{
		registryClient: registryClient,
	}
}

// DetectFramework uses the alizer library in order to detect the devfile
// to use depending on the files in the path
func (o *Alizer) DetectFramework(path string) (recognizer.DevFileType, registry.Registry, error) {
	types := []recognizer.DevFileType{}
	components, err := o.registryClient.ListDevfileStacks("")
	if err != nil {
		return recognizer.DevFileType{}, registry.Registry{}, err
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
		return recognizer.DevFileType{}, registry.Registry{}, err
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

func GetDevfileLocationFromDetection(typ recognizer.DevFileType, registry registry.Registry) *DevfileLocation {
	return &DevfileLocation{
		Devfile:         typ.Name,
		DevfileRegistry: registry.Name,
	}
}
