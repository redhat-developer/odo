package libdevfile

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
)

// imageComponent implements the component interface
type imageComponent struct {
	component  v1alpha2.Component
	devfileObj parser.DevfileObj
}

var _ component = (*imageComponent)(nil)

func newImageComponent(devfileObj parser.DevfileObj, component v1alpha2.Component) *imageComponent {
	return &imageComponent{
		component:  component,
		devfileObj: devfileObj,
	}
}

func (e *imageComponent) CheckValidity() error {
	return nil
}

func (e *imageComponent) Apply(handler Handler, kind v1alpha2.CommandGroupKind) error {
	return handler.ApplyImage(e.component)
}

// GetImageComponentsToPushAutomatically returns the list of Image components that can be automatically created on startup.
// The list returned is governed by the AutoBuild field in each component.
// All components with AutoBuild set to true are included, along with those with no AutoBuild set and not-referenced.
func GetImageComponentsToPushAutomatically(devfileObj parser.DevfileObj) ([]v1alpha2.Component, error) {
	imageComponents, err := devfileObj.Data.GetComponents(parsercommon.DevfileOptions{
		ComponentOptions: parsercommon.ComponentOptions{ComponentType: v1alpha2.ImageComponentType},
	})
	if err != nil {
		return nil, err
	}

	allApplyCommands, err := devfileObj.Data.GetCommands(parsercommon.DevfileOptions{
		CommandOptions: parsercommon.CommandOptions{CommandType: v1alpha2.ApplyCommandType},
	})
	if err != nil {
		return nil, err
	}

	m := make(map[string]v1alpha2.Component)
	for _, comp := range imageComponents {
		var add bool
		if comp.Image.AutoBuild == nil {
			// auto-created only if not referenced by any apply command
			if !isComponentReferenced(allApplyCommands, comp.Name) {
				add = true
			}
		} else if *comp.Image.AutoBuild {
			add = true
		}
		if !add {
			continue
		}
		if _, present := m[comp.Name]; !present {
			m[comp.Name] = comp
		}
	}

	var result []v1alpha2.Component
	for _, comp := range m {
		result = append(result, comp)
	}
	return result, nil
}
