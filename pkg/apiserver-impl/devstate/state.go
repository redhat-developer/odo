package devstate

import (
	"fmt"
	"strings"

	apidevfile "github.com/devfile/api/v2/pkg/devfile"
	"github.com/devfile/library/v2/pkg/devfile"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	context "github.com/devfile/library/v2/pkg/devfile/parser/context"
	"github.com/devfile/library/v2/pkg/testingutil/filesystem"

	. "github.com/redhat-developer/odo/pkg/apiserver-gen/go"

	"k8s.io/utils/pointer"
)

type DevfileState struct {
	Devfile parser.DevfileObj
	FS      filesystem.Filesystem
}

func NewDevfileState() DevfileState {
	s := DevfileState{
		FS: filesystem.NewFakeFs(),
	}
	// this should never fail, as the parameters are constant
	_, _ = s.SetDevfileContent(`schemaVersion: 2.2.0`)
	return s
}

// SetDevfileContent replaces the devfile with a new content
// If an error occurs, the Devfile is not modified
func (o *DevfileState) SetDevfileContent(content string) (DevfileContent, error) {
	parserArgs := parser.ParserArgs{
		Data:                          []byte(content),
		ConvertKubernetesContentInUri: pointer.Bool(false),
		SetBooleanDefaults:            pointer.Bool(false),
	}
	var err error
	devfile, _, err := devfile.ParseDevfileAndValidate(parserArgs)
	if err != nil {
		return DevfileContent{}, fmt.Errorf("error parsing devfile YAML: %w", err)
	}
	o.Devfile = devfile
	o.Devfile.Ctx = context.FakeContext(o.FS, o.Devfile.Ctx.GetAbsPath())
	return o.GetContent()
}

func (o *DevfileState) SetMetadata(
	name string,
	version string,
	displayName string,
	description string,
	tags string,
	architectures string,
	icon string,
	globalMemoryLimit string,
	projectType string,
	language string,
	website string,
	provider string,
	supportUrl string,
) (DevfileContent, error) {
	o.Devfile.Data.SetMetadata(apidevfile.DevfileMetadata{
		Name:              name,
		Version:           version,
		DisplayName:       displayName,
		Description:       description,
		Tags:              splitTags(tags),
		Architectures:     splitArchitectures(architectures),
		Icon:              icon,
		GlobalMemoryLimit: globalMemoryLimit,
		ProjectType:       projectType,
		Language:          language,
		Website:           website,
		Provider:          provider,
		SupportUrl:        supportUrl,
	})
	return o.GetContent()
}
func splitArchitectures(architectures string) []apidevfile.Architecture {
	if architectures == "" {
		return nil
	}
	parts := strings.Split(architectures, SEPARATOR)
	result := make([]apidevfile.Architecture, len(parts))
	for i, arch := range parts {
		result[i] = apidevfile.Architecture(strings.Trim(arch, " "))
	}
	return result
}

func splitTags(tags string) []string {
	if tags == "" {
		return nil
	}
	parts := strings.Split(tags, SEPARATOR)
	result := make([]string, len(parts))
	for i, tag := range parts {
		result[i] = strings.Trim(tag, " ")
	}
	return result
}
