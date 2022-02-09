package v2

import (
	"fmt"
	"reflect"
	"strings"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

// GetComponents returns the slice of Component objects parsed from the Devfile
func (d *DevfileV2) GetComponents(options common.DevfileOptions) ([]v1.Component, error) {

	if reflect.DeepEqual(options, common.DevfileOptions{}) {
		return d.Components, nil
	}

	var components []v1.Component
	for _, component := range d.Components {
		// Filter Component Attributes
		filterIn, err := common.FilterDevfileObject(component.Attributes, options)
		if err != nil {
			return nil, err
		} else if !filterIn {
			continue
		}

		// Filter Component Type - Container, Volume, etc.
		componentType, err := common.GetComponentType(component)
		if err != nil {
			return nil, err
		}
		if options.ComponentOptions.ComponentType != "" && componentType != options.ComponentOptions.ComponentType {
			continue
		}

		if options.FilterByName == "" || component.Name == options.FilterByName {
			components = append(components, component)
		}
	}

	return components, nil
}

// GetDevfileContainerComponents iterates through the components in the devfile and returns a list of devfile container components.
// Deprecated, use GetComponents() with the DevfileOptions.
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

// GetDevfileVolumeComponents iterates through the components in the devfile and returns a list of devfile volume components.
// Deprecated, use GetComponents() with the DevfileOptions.
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
// a component is considered as invalid if it is already defined
// component list passed in will be all processed, and returns a total error of all invalid components
func (d *DevfileV2) AddComponents(components []v1.Component) error {
	var errorsList []string
	for _, component := range components {
		var err error
		for _, devfileComponent := range d.Components {
			if component.Name == devfileComponent.Name {
				err = &common.FieldAlreadyExistError{Name: component.Name, Field: "component"}
				errorsList = append(errorsList, err.Error())
				break
			}
		}
		if err == nil {
			d.Components = append(d.Components, component)
		}
	}
	if len(errorsList) > 0 {
		return fmt.Errorf("errors while adding components:\n%s", strings.Join(errorsList, "\n"))
	}
	return nil
}

// UpdateComponent updates the component with the given name
// return an error if the component is not found
func (d *DevfileV2) UpdateComponent(component v1.Component) error {
	for i := range d.Components {
		if d.Components[i].Name == component.Name {
			d.Components[i] = component
			return nil
		}
	}
	return fmt.Errorf("update component failed: component %s not found", component.Name)
}

// DeleteComponent removes the specified component
func (d *DevfileV2) DeleteComponent(name string) error {

	for i := range d.Components {
		if d.Components[i].Name == name {
			d.Components = append(d.Components[:i], d.Components[i+1:]...)
			return nil
		}
	}

	return &common.FieldNotFoundError{
		Field: "component",
		Name:  name,
	}
}
