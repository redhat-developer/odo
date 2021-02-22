package v2

import (
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

// GetComponents returns the slice of Component objects parsed from the Devfile
func (d *DevfileV2) GetComponents(options common.DevfileOptions) ([]v1.Component, error) {
	if len(options.Filter) == 0 {
		return d.Components, nil
	}

	var components []v1.Component
	for _, comp := range d.Components {
		filterIn, err := common.FilterDevfileObject(comp.Attributes, options)
		if err != nil {
			return nil, err
		}

		if filterIn {
			components = append(components, comp)
		}
	}

	return components, nil
}

// GetDevfileContainerComponents iterates through the components in the devfile and returns a list of devfile container components
func (d *DevfileV2) GetDevfileContainerComponents(options common.DevfileOptions) ([]v1.Component, error) {
	var components []v1.Component
	devfileComponents, err := d.GetComponents(options)
	if err != nil {
		return nil, err
	}
	for _, comp := range devfileComponents {
		if comp.Container != nil {
			components = append(components, comp)
		}
	}
	return components, nil
}

// GetDevfileVolumeComponents iterates through the components in the devfile and returns a list of devfile volume components
func (d *DevfileV2) GetDevfileVolumeComponents(options common.DevfileOptions) ([]v1.Component, error) {
	var components []v1.Component
	devfileComponents, err := d.GetComponents(options)
	if err != nil {
		return nil, err
	}
	for _, comp := range devfileComponents {
		if comp.Volume != nil {
			components = append(components, comp)
		}
	}
	return components, nil
}

// AddComponents adds the slice of Component objects to the devfile's components
// if a component is already defined, error out
func (d *DevfileV2) AddComponents(components []v1.Component) error {

	componentMap := make(map[string]bool)

	for _, component := range d.Components {
		componentMap[component.Name] = true
	}
	for _, component := range components {
		if _, ok := componentMap[component.Name]; !ok {
			d.Components = append(d.Components, component)
		} else {
			return &common.FieldAlreadyExistError{Name: component.Name, Field: "component"}
		}
	}
	return nil
}

// UpdateComponent updates the component with the given name
func (d *DevfileV2) UpdateComponent(component v1.Component) {
	index := -1
	for i := range d.Components {
		if d.Components[i].Name == component.Name {
			index = i
			break
		}
	}
	if index != -1 {
		d.Components[index] = component
	}
}
