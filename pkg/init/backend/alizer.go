package backend

import (
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/alizer/go/pkg/apis/language"
	"github.com/redhat-developer/alizer/go/pkg/apis/recognizer"
	"github.com/redhat-developer/odo/pkg/init/asker"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

type AlizerBackend struct {
	asker asker.Asker
}

func NewAlizerBackend(asker asker.Asker) *AlizerBackend {
	return &AlizerBackend{
		asker: asker,
	}
}

func (o *AlizerBackend) Validate(flags map[string]string, fs filesystem.Filesystem, dir string) error {
	return nil
}

// DetectFramework uses the anlizer library in order to detect the language
// as well as framework and tool of a Devfile location
// Passes in PATH and gets a function that contains strings describing the
// directory's detected languages.
func detectFramework(path string) ([]language.Language, error) {
	return recognizer.Analyze(path)
}

func (o *AlizerBackend) SelectDevfile(flags map[string]string, fs filesystem.Filesystem, dir string) (location *DevfileLocation, err error) {
	aliases := map[string]string{
		"JavaScript": "nodejs",
	}

	languages, err := detectFramework(dir)
	if err != nil {
		return nil, err
	}
	if len(languages) == 0 {
		return nil, fmt.Errorf("no language detected")
	}

	fmt.Printf("Based on the files in the current directory odo detected\nLanguage: %s\nProject type: %s\n", languages[0].Name, languages[0].Tools[0])

	fmt.Printf("%+v\n", languages[0])
	var lang string
	var ok bool
	if lang, ok = aliases[languages[0].Name]; !ok {
		return nil, fmt.Errorf("unable to match devfile for detected language %q", languages[0].Name)
	}

	confirm, err := o.asker.AskCorrect()
	if err != nil {
		return nil, err
	}
	if !confirm {
		return nil, nil
	}
	return &DevfileLocation{
		Devfile: lang,
	}, nil
}

func (o *AlizerBackend) SelectStarterProject(devfile parser.DevfileObj, flags map[string]string) (starter *v1alpha2.StarterProject, err error) {
	return nil, nil
}

func (o *AlizerBackend) PersonalizeName(devfile parser.DevfileObj, flags map[string]string) error {
	return nil
}
