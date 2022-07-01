package alizer

import (
	"github.com/redhat-developer/alizer/go/pkg/apis/recognizer"
	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/registry"
)

type Alizer struct {
	registryClient registry.Client
}

var _ Client = (*Alizer)(nil)

func NewAlizerClient(registryClient registry.Client) *Alizer {
	return &Alizer{
		registryClient: registryClient,
	}
}

// DetectFramework uses the alizer library in order to detect the devfile
// to use depending on the files in the path
func (o *Alizer) DetectFramework(path string) (recognizer.DevFileType, api.Registry, error) {
	types := []recognizer.DevFileType{}
	components, err := o.registryClient.ListDevfileStacks("", "", "", false)
	if err != nil {
		return recognizer.DevFileType{}, api.Registry{}, err
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
		return recognizer.DevFileType{}, api.Registry{}, err
	}
	return types[typ], components.Items[typ].Registry, nil
}

func GetDevfileLocationFromDetection(typ recognizer.DevFileType, registry api.Registry) *api.DevfileLocation {
	return &api.DevfileLocation{
		Devfile:         typ.Name,
		DevfileRegistry: registry.Name,
	}
}
