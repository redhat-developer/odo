package overriding

import (
	"fmt"
	"reflect"
	"strings"

	dw "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// MergeDevWorkspaceTemplateSpec implements the merging logic of a main devfile content with flattened, already-overridden parent devfiles or plugins.
// On a `main` `DevWorkspaceTemplateSpec` (which is the core part of a devfile, without the `apiVersion` and `metadata`),
// it allows adding all the new overridden elements provided by flattened parent and plugins
//
// Returns non-nil error if there are duplicate (== with same key) commands, components or projects between the
// main content and the parent or plugins.
//
// The result is a transformed `DevWorkspaceTemplateSpec` object, that does not contain any `plugin` component
// (since they are expected to be provided as flattened overridden devfiles in the arguments)
func MergeDevWorkspaceTemplateSpec(
	mainContent *dw.DevWorkspaceTemplateSpecContent,
	parentFlattenedContent *dw.DevWorkspaceTemplateSpecContent,
	pluginFlattenedContents ...*dw.DevWorkspaceTemplateSpecContent) (*dw.DevWorkspaceTemplateSpecContent, error) {

	allContents := []*dw.DevWorkspaceTemplateSpecContent{}
	if parentFlattenedContent != nil {
		allContents = append(allContents, parentFlattenedContent)
	}
	if len(pluginFlattenedContents) > 0 {
		allContents = append(allContents, pluginFlattenedContents...)
	}
	allContents = append(allContents, mainContent)

	// Check for conflicts
	if parentFlattenedContent != nil {
		if err := ensureNoConflictWithParent(mainContent, parentFlattenedContent); err != nil {
			return nil, err
		}
	}
	if len(pluginFlattenedContents) > 0 {
		if err := ensureNoConflictsWithPlugins(mainContent, pluginFlattenedContents...); err != nil {
			return nil, err
		}
		if parentFlattenedContent != nil {
			// also need to ensure no conflict between parent and plugins
			if err := ensureNoConflictsWithPlugins(parentFlattenedContent, pluginFlattenedContents...); err != nil {
				return nil, err
			}
		}
	}

	result := dw.DevWorkspaceTemplateSpecContent{}

	// Merge top-level lists (Commands, Projects, Components, etc ...)

	topLevelListsNames := result.GetToplevelLists()
	topLevelListsByContent := []dw.TopLevelLists{}
	for _, content := range allContents {
		topLevelListsByContent = append(topLevelListsByContent, content.GetToplevelLists())
	}

	resultValue := reflect.ValueOf(&result).Elem()
	for toplevelListName := range topLevelListsNames {
		listType, fieldExists := resultValue.Type().FieldByName(toplevelListName)
		if !fieldExists {
			return nil, fmt.Errorf("field '%v' is unknown in %v struct (this should never happen)", toplevelListName, reflect.TypeOf(resultValue).Name())
		}
		if listType.Type.Kind() != reflect.Slice {
			return nil, fmt.Errorf("field '%v' in %v struct is not a slice (this should never happen)", toplevelListName, reflect.TypeOf(resultValue).Name())
		}
		listElementType := listType.Type.Elem()
		resultToplevelListValue := resultValue.FieldByName(toplevelListName)
		for contentIndex, content := range allContents {
			toplevelLists := topLevelListsByContent[contentIndex]
			keyedList := toplevelLists[toplevelListName]
			for _, keyed := range keyedList {
				if content == mainContent {
					if component, isComponent := keyed.(dw.Component); isComponent &&
						component.Plugin != nil {
						continue
					}
				}
				if resultToplevelListValue.IsNil() {
					resultToplevelListValue.Set(reflect.MakeSlice(reflect.SliceOf(listElementType), 0, len(keyedList)))
				}
				resultToplevelListValue.Set(reflect.Append(resultToplevelListValue, reflect.ValueOf(keyed)))
			}
		}
	}

	preStartCommands := sets.String{}
	postStartCommands := sets.String{}
	preStopCommands := sets.String{}
	postStopCommands := sets.String{}
	for _, content := range allContents {
		if content.Events != nil {
			if result.Events == nil {
				result.Events = &dw.Events{}
			}
			preStartCommands = preStartCommands.Union(sets.NewString(content.Events.PreStart...))
			postStartCommands = postStartCommands.Union(sets.NewString(content.Events.PostStart...))
			preStopCommands = preStopCommands.Union(sets.NewString(content.Events.PreStop...))
			postStopCommands = postStopCommands.Union(sets.NewString(content.Events.PostStop...))
		}

		if len(content.Variables) > 0 {
			if len(result.Variables) == 0 {
				result.Variables = make(map[string]string)
			}
			for k, v := range content.Variables {
				result.Variables[k] = v
			}
		}

		var err error
		if len(content.Attributes) > 0 {
			if len(result.Attributes) == 0 {
				result.Attributes = attributes.Attributes{}
			}
			for k, v := range content.Attributes {
				result.Attributes.Put(k, v, &err)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	if result.Events != nil {
		result.Events.PreStart = preStartCommands.List()
		result.Events.PostStart = postStartCommands.List()
		result.Events.PreStop = preStopCommands.List()
		result.Events.PostStop = postStopCommands.List()
	}

	return &result, nil
}

// MergeDevWorkspaceTemplateSpecBytes implements the merging logic of a main devfile content with flattened, already-overridden parent devfiles or plugins.
// On an json or yaml document that contains the core content of the devfile (which is the core part of a devfile, without the `apiVersion` and `metadata`),
// it allows adding all the new overridden elements provided by flattened parent and plugins (also provided as json or yaml documents)
//
// It is not allowed for to have duplicate (== with same key) commands, components or projects between the main content and the parent or plugins.
// An error would be thrown
//
// The result is a transformed `DevfileWorkspaceTemplateSpec` object, that does not contain any `plugin` component
// (since they are expected to be provided as flattened overridden devfiles in the arguments)
func MergeDevWorkspaceTemplateSpecBytes(originalBytes []byte, flattenedParentBytes []byte, flattenPluginsBytes ...[]byte) (*dw.DevWorkspaceTemplateSpecContent, error) {
	originalJson, err := yaml.ToJSON(originalBytes)
	if err != nil {
		return nil, err
	}

	original := dw.DevWorkspaceTemplateSpecContent{}
	err = json.Unmarshal(originalJson, &original)
	if err != nil {
		return nil, err
	}

	flattenedParentJson, err := yaml.ToJSON(flattenedParentBytes)
	if err != nil {
		return nil, err
	}

	flattenedParent := dw.DevWorkspaceTemplateSpecContent{}
	err = json.Unmarshal(flattenedParentJson, &flattenedParent)
	if err != nil {
		return nil, err
	}

	flattenedPlugins := []*dw.DevWorkspaceTemplateSpecContent{}
	for _, flattenedPluginBytes := range flattenPluginsBytes {
		flattenedPluginJson, err := yaml.ToJSON(flattenedPluginBytes)
		if err != nil {
			return nil, err
		}

		flattenedPlugin := dw.DevWorkspaceTemplateSpecContent{}
		err = json.Unmarshal(flattenedPluginJson, &flattenedPlugin)
		if err != nil {
			return nil, err
		}
		flattenedPlugins = append(flattenedPlugins, &flattenedPlugin)
	}

	return MergeDevWorkspaceTemplateSpec(&original, &flattenedParent, flattenedPlugins...)
}

func ensureNoConflictWithParent(mainContent *dw.DevWorkspaceTemplateSpecContent, parentflattenedContent *dw.DevWorkspaceTemplateSpecContent) error {
	return checkKeys(func(elementType string, keysSets []sets.String) []error {
		mainKeys := keysSets[0]
		parentOrPluginKeys := keysSets[1]
		overriddenElementsInMainContent := mainKeys.Intersection(parentOrPluginKeys)
		if overriddenElementsInMainContent.Len() > 0 {
			return []error{fmt.Errorf("Some %s are already defined in parent: %s. "+
				"If you want to override them, you should do it in the parent scope.",
				elementType,
				strings.Join(overriddenElementsInMainContent.List(), ", "))}
		}
		return []error{}
	},
		mainContent, parentflattenedContent)
}

func ensureNoConflictsWithPlugins(mainContent *dw.DevWorkspaceTemplateSpecContent, pluginFlattenedContents ...*dw.DevWorkspaceTemplateSpecContent) error {
	getPluginKey := func(pluginIndex int) string {
		index := 0
		for _, comp := range mainContent.Components {
			if comp.Plugin != nil {
				if pluginIndex == index {
					return comp.Name
				}
				index++
			}
		}
		return "unknown"
	}

	allSpecs := []dw.TopLevelListContainer{mainContent}
	for _, pluginFlattenedContent := range pluginFlattenedContents {
		allSpecs = append(allSpecs, pluginFlattenedContent)
	}
	return checkKeys(func(elementType string, keysSets []sets.String) []error {
		mainKeys := keysSets[0]
		pluginKeysSets := keysSets[1:]
		errs := []error{}
		for pluginNumber, pluginKeys := range pluginKeysSets {
			overriddenElementsInMainContent := mainKeys.Intersection(pluginKeys)

			if overriddenElementsInMainContent.Len() > 0 {
				errs = append(errs, fmt.Errorf("Some %s are already defined in plugin '%s': %s. "+
					"If you want to override them, you should do it in the plugin scope.",
					elementType,
					getPluginKey(pluginNumber),
					strings.Join(overriddenElementsInMainContent.List(), ", ")))
			}
		}
		return errs
	},
		allSpecs...)
}
