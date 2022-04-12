package backend

import (
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/alizer"
	"github.com/redhat-developer/odo/pkg/init/asker"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

type AlizerBackend struct {
	askerClient  asker.Asker
	alizerClient alizer.Client
}

func NewAlizerBackend(askerClient asker.Asker, alizerClient alizer.Client) *AlizerBackend {
	return &AlizerBackend{
		askerClient:  askerClient,
		alizerClient: alizerClient,
	}
}

func (o *AlizerBackend) Validate(flags map[string]string, fs filesystem.Filesystem, dir string) error {
	return nil
}

// SelectDevfile calls thz Alizer to detect the devfile and asks for confirmation to the user
func (o *AlizerBackend) SelectDevfile(flags map[string]string, fs filesystem.Filesystem, dir string) (location *alizer.DevfileLocation, err error) {
	selected, registry, err := o.alizerClient.DetectFramework(dir)
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
	return alizer.GetDevfileLocationFromDetection(selected, registry), nil
}

func (o *AlizerBackend) SelectStarterProject(devfile parser.DevfileObj, flags map[string]string) (starter *v1alpha2.StarterProject, err error) {
	return nil, nil
}

func (o *AlizerBackend) PersonalizeName(devfile parser.DevfileObj, flags map[string]string) (string, error) {
	return devfile.GetMetadataName(), nil
}

func (o *AlizerBackend) PersonalizeDevfileConfig(devfile parser.DevfileObj) (parser.DevfileObj, error) {
	return devfile, nil
}
