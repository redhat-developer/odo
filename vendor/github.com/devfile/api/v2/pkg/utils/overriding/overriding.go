package overriding

import (
	"fmt"
	"reflect"
	"strings"

	workspaces "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	unions "github.com/devfile/api/v2/pkg/utils/unions"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/sets"
	strategicpatch "k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// OverrideDevWorkspaceTemplateSpecBytes implements the overriding logic for parent devfiles or plugins.
// On a json or yaml document that contains the core content of the devfile (without the `apiVersion` and `metadata`),
// it allows applying a `patch` which is a document fragment of the same schema.
//
// The Overriding logic is implemented according to strategic merge patch rules, as defined here:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md#background
//
// The result is a transformed `DevfileWorkspaceTemplateSpec` object that can be serialized back to yaml or json.
func OverrideDevWorkspaceTemplateSpecBytes(originalBytes []byte, patchBytes []byte) (*workspaces.DevWorkspaceTemplateSpecContent, error) {
	originalJson, err := yaml.ToJSON(originalBytes)
	if err != nil {
		return nil, err
	}

	original := workspaces.DevWorkspaceTemplateSpecContent{}
	err = json.Unmarshal(originalJson, &original)
	if err != nil {
		return nil, err
	}

	patchJson, err := yaml.ToJSON(patchBytes)
	if err != nil {
		return nil, err
	}

	patch := workspaces.ParentOverrides{}
	err = json.Unmarshal(patchJson, &patch)
	if err != nil {
		return nil, err
	}

	return OverrideDevWorkspaceTemplateSpec(&original, &patch)

}

// OverrideDevWorkspaceTemplateSpec implements the overriding logic for parent devfiles or plugins.
// On an `original` `DevfileWorkspaceTemplateSpec` (which is the core part of a devfile, without the `apiVersion` and `metadata`),
// it allows applying a `patch` which is a `ParentOverrides` or a `PluginOverrides` object.
//
// The Overriding logic is implemented according to strategic merge patch rules, as defined here:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md#background
//
// The result is a transformed `DevfileWorkspaceTemplateSpec` object.
func OverrideDevWorkspaceTemplateSpec(original *workspaces.DevWorkspaceTemplateSpecContent, patch workspaces.Overrides) (*workspaces.DevWorkspaceTemplateSpecContent, error) {
	if err := ensureOnlyExistingElementsAreOverridden(original, patch); err != nil {
		return nil, err
	}

	if err := unions.Normalize(&original); err != nil {
		return nil, err
	}
	if err := unions.Normalize(&patch); err != nil {
		return nil, err
	}

	normalizedOriginalBytes, err := json.Marshal(original)
	if err != nil {
		return nil, err
	}

	originalMap, err := handleUnmarshal(normalizedOriginalBytes)
	if err != nil {
		return nil, err
	}

	normalizedPatchBytes, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}
	patchMap, err := handleUnmarshal(normalizedPatchBytes)
	if err != nil {
		return nil, err
	}

	schema, err := strategicpatch.NewPatchMetaFromStruct(original)
	if err != nil {
		return nil, err
	}

	patchedMap, err := strategicpatch.StrategicMergeMapPatchUsingLookupPatchMeta(originalMap, patchMap, mapEnabledPatchMetaFromStruct{schema})
	if err != nil {
		return nil, err
	}

	patchedBytes, err := json.Marshal(patchedMap)
	if err != nil {
		return nil, err
	}

	patched := workspaces.DevWorkspaceTemplateSpecContent{}
	err = json.Unmarshal(patchedBytes, &patched)
	if err != nil {
		return nil, err
	}

	if err = unions.Simplify(&patched); err != nil {
		return nil, err
	}
	return &patched, nil
}

func ensureOnlyExistingElementsAreOverridden(spec *workspaces.DevWorkspaceTemplateSpecContent, overrides workspaces.Overrides) error {
	return checkKeys(func(elementType string, keysSets []sets.String) []error {
		if len(keysSets) <= 1 {
			return []error{}
		}
		specKeys := keysSets[0]
		overlayKeys := keysSets[1]
		newElementsInOverlay := overlayKeys.Difference(specKeys)
		if newElementsInOverlay.Len() > 0 {
			return []error{fmt.Errorf("Some %s do not override any existing element: %s. "+
				"They should be defined in the main body, as new elements, not in the overriding section",
				elementType,
				strings.Join(newElementsInOverlay.List(), ", "))}
		}
		return []error{}
	},
		spec, overrides)
}

type mapEnabledPatchMetaFromStruct struct {
	delegate strategicpatch.PatchMetaFromStruct
}

var _ strategicpatch.LookupPatchMeta = (*mapEnabledPatchMetaFromStruct)(nil)

func (s *mapEnabledPatchMetaFromStruct) replaceMapWithSingleKeyStruct(key string, elementIsSlice bool) {
	t := s.delegate.T
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() == reflect.Map {
		typeName := t.Name()
		tag := `json:"` + key + `"`
		elemType := t.Elem()
		if typeName == "Attributes" {
			if elementIsSlice {
				elemType = reflect.SliceOf(reflect.StructOf([]reflect.StructField{}))
			} else {
				tag = tag + ` patchStrategy:"replace"`
				elemType = reflect.StructOf([]reflect.StructField{})
			}
		}

		s.delegate.T = reflect.StructOf([]reflect.StructField{{Name: strings.Title(key), Type: elemType, Tag: reflect.StructTag(tag)}})
	}
}

func (s mapEnabledPatchMetaFromStruct) LookupPatchMetadataForStruct(key string) (strategicpatch.LookupPatchMeta, strategicpatch.PatchMeta, error) {
	s.replaceMapWithSingleKeyStruct(key, false)

	schema, patchMeta, err := s.delegate.LookupPatchMetadataForStruct(key)
	var lookupPatchMeta strategicpatch.LookupPatchMeta = nil
	if schema != nil {
		lookupPatchMeta = mapEnabledPatchMetaFromStruct{delegate: schema.(strategicpatch.PatchMetaFromStruct)}
	}
	return lookupPatchMeta, patchMeta, err
}

func (s mapEnabledPatchMetaFromStruct) LookupPatchMetadataForSlice(key string) (strategicpatch.LookupPatchMeta, strategicpatch.PatchMeta, error) {
	s.replaceMapWithSingleKeyStruct(key, true)

	schema, patchMeta, err := s.delegate.LookupPatchMetadataForSlice(key)
	var lookupPatchMeta strategicpatch.LookupPatchMeta = nil
	if schema != nil {
		lookupPatchMeta = mapEnabledPatchMetaFromStruct{schema.(strategicpatch.PatchMetaFromStruct)}
	}
	return lookupPatchMeta, patchMeta, err
}

func (s mapEnabledPatchMetaFromStruct) Name() string {
	return s.delegate.Name()
}
