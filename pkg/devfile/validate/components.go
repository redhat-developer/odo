package validate

import (
	"fmt"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// Errors
var (
	ErrorNoComponents              = "no components present"
	ErrorNoContainerComponent      = fmt.Sprintf("odo requires atleast one component of type '%s' in devfile", common.ContainerComponentType)
	ErrorDuplicateVolumeComponents = "duplicate volume components present in devfile"
)

// validateComponents validates all the devfile components
func validateComponents(components []common.DevfileComponent) error {

	// components cannot be empty
	if len(components) < 1 {
		return fmt.Errorf(ErrorNoComponents)
	}

	processedVolumes := make(map[string]bool)
	// var containerVolumeMountNames []string

	// Check if component of type container is present
	// and if volume components are unique
	isContainerComponentPresent := false
	for _, component := range components {
		if component.Container != nil {
			isContainerComponentPresent = true
		}

		if component.Volume != nil {
			if _, ok := processedVolumes[component.Volume.Name]; !ok {
				processedVolumes[component.Volume.Name] = true
			} else {
				return fmt.Errorf(ErrorDuplicateVolumeComponents)
			}
		}
	}

	if !isContainerComponentPresent {
		return fmt.Errorf(ErrorNoContainerComponent)
	}

	// Successful
	return nil
}
