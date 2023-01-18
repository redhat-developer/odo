package libdevfile

import (
	"errors"
	"net/url"
	"sort"
	"strings"
	
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	"github.com/devfile/api/v2/pkg/validation"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
)

const _importSourceAttributeUriPrefix = "uri: "

// GetReferencedLocalFiles returns the local files referenced by the Devfile. This includes:
// - the non-inlined Kubernetes and Openshift components
// - the Dockerfiles of Image components
// - the parent devfile
// - resursively, the local files referenced by the parent Devfile
// The passed Devfile must be flattened
func GetReferencedLocalFiles(devfileObj parser.DevfileObj) (result []string, err error) {

	setResult := map[string]struct{}{}

	parent := devfileObj.Data.GetParent()
	if parent != nil {
		return nil, errors.New("devfile must be flattened")
	}

	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return nil, err
	}

	for _, component := range components {
		var componentType v1alpha2.ComponentType
		componentType, err = common.GetComponentType(component)
		if err != nil {
			return nil, err
		}

		switch componentType {
		case v1alpha2.KubernetesComponentType:
			setResult, err = appendUriIfFile(setResult, component.Kubernetes.Uri)
			if err != nil {
				return nil, err
			}

		case v1alpha2.OpenshiftComponentType:
			setResult, err = appendUriIfFile(setResult, component.Openshift.Uri)
			if err != nil {
				return nil, err
			}

		case v1alpha2.ImageComponentType:
			if component.Image.Dockerfile != nil {
				setResult, err = appendUriIfFile(setResult, component.Image.Dockerfile.Uri)
				if err != nil {
					return nil, err
				}
			}
		}

		setResult, err = getFromAttributes(setResult, component.Attributes)
		if err != nil {
			return nil, err
		}
	}

	commands, err := devfileObj.Data.GetCommands(common.DevfileOptions{})
	if err != nil {
		return nil, err
	}
	for _, command := range commands {
		setResult, err = getFromAttributes(setResult, command.Attributes)
		if err != nil {
			return nil, err
		}
	}

	result = make([]string, 0, len(setResult))
	for k := range setResult {
		result = append(result, k)
	}
	sort.Strings(result)
	return result, nil
}

// appendUriIfFile appends uri to the result if the uri is a local path
func appendUriIfFile(result map[string]struct{}, uri string) (map[string]struct{}, error) {
	if uri != "" {
		u, err := url.Parse(uri)
		if err != nil {
			return nil, err
		}
		if u.Scheme == "" {
			result[uri] = struct{}{}
		}
	}
	return result, nil
}

// getFromAttributes extracts paths from attributes entries with key "api.devfile.io/imported-from"
// containing a uri reference as a local path
func getFromAttributes(result map[string]struct{}, attributes attributes.Attributes) (map[string]struct{}, error) {
	if val, ok := attributes[validation.ImportSourceAttribute]; ok {
		strVal := string(val.Raw)
		strVal = strings.Trim(strVal, `"`)
		if strings.HasPrefix(strVal, _importSourceAttributeUriPrefix) {
			parentUri := strings.TrimLeft(strVal, _importSourceAttributeUriPrefix)
			var err error
			result, err = appendUriIfFile(result, parentUri)
			if err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}
