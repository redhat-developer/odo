package exports

import (
	"strings"
	"syscall/js"

	apidevfile "github.com/devfile/api/v2/pkg/devfile"

	"github.com/feloy/devfile-builder/wasm/pkg/global"
	"github.com/feloy/devfile-builder/wasm/pkg/utils"
)

// setMetadata

func SetMetadataWrapper(this js.Value, args []js.Value) interface{} {
	return result(
		setMetadata(args[0]),
	)
}

func setMetadata(metadata js.Value) (map[string]interface{}, error) {
	global.Devfile.Data.SetMetadata(apidevfile.DevfileMetadata{
		Name:              metadata.Get("name").String(),
		Version:           metadata.Get("version").String(),
		DisplayName:       metadata.Get("displayName").String(),
		Description:       metadata.Get("description").String(),
		Tags:              splitTags(metadata.Get("tags").String()),
		Architectures:     splitArchitectures(metadata.Get("architectures").String()),
		Icon:              metadata.Get("icon").String(),
		GlobalMemoryLimit: metadata.Get("globalMemoryLimit").String(),
		ProjectType:       metadata.Get("projectType").String(),
		Language:          metadata.Get("language").String(),
		Website:           metadata.Get("website").String(),
		Provider:          metadata.Get("provider").String(),
		SupportUrl:        metadata.Get("supportUrl").String(),
	})
	return utils.GetContent()
}

func splitArchitectures(architectures string) []apidevfile.Architecture {
	if architectures == "" {
		return nil
	}
	parts := strings.Split(architectures, utils.SEPARATOR)
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
	parts := strings.Split(tags, utils.SEPARATOR)
	result := make([]string, len(parts))
	for i, tag := range parts {
		result[i] = strings.Trim(tag, " ")
	}
	return result
}
