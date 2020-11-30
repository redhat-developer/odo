package v2

import (
	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

// GetComponents returns the slice of Component objects parsed from the Devfile
func (d *DevfileV2) GetComponents() []v1.Component {
	return d.Components
}

// GetDevfileContainerComponents iterates through the components in the devfile and returns a list of devfile container components
func (d *DevfileV2) GetDevfileContainerComponents() []v1.Component {
	var components []v1.Component
	for _, comp := range d.GetComponents() {
		if comp.Container != nil {
			components = append(components, comp)
		}
	}
	return components
}

// GetDevfileVolumeComponents iterates through the components in the devfile and returns a list of devfile volume components
func (d *DevfileV2) GetDevfileVolumeComponents() []v1.Component {
	var components []v1.Component
	for _, comp := range d.GetComponents() {
		if comp.Volume != nil {
			components = append(components, comp)
		}
	}
	return components
}

// AddComponents adds the slice of Component objects to the devfile's components
// if a component is already defined, error out
func (d *DevfileV2) AddComponents(components []v1.Component) error {

	// different map for volume and container component as a volume and a container with same name
	// can exist in devfile
	containerMap := make(map[string]bool)
	volumeMap := make(map[string]bool)

	for _, component := range d.Components {
		if component.Volume != nil {
			volumeMap[component.Name] = true
		}
		if component.Container != nil {
			containerMap[component.Name] = true
		}
	}

	for _, component := range components {

		if component.Volume != nil {
			if _, ok := volumeMap[component.Name]; !ok {
				d.Components = append(d.Components, component)
			} else {
				return &common.FieldAlreadyExistError{Name: component.Name, Field: "component"}
			}
		}

		if component.Container != nil {
			if _, ok := containerMap[component.Name]; !ok {
				d.Components = append(d.Components, component)
			} else {
				return &common.FieldAlreadyExistError{Name: component.Name, Field: "component"}
			}
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
