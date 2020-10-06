package validate

import (
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	genericValidation "github.com/openshift/odo/pkg/devfile/validate/generic"
)

// validateComponents validates the devfile components:
// 1. there should be at least one component
// 2. there should be at least one container component
// 3. the components should be a valid devfile component
func validateComponents(components []common.DevfileComponent) error {

	// components cannot be empty
	if len(components) < 1 {
		return &NoComponentsError{}
	}

	// Check if component of type container is present
	// and if volume components are unique
	isContainerComponentPresent := false
	for _, component := range components {
		if component.IsContainer() {
			isContainerComponentPresent = true
		}
	}

	if !isContainerComponentPresent {
		return &NoContainerComponentError{}
	}

	err := genericValidation.ValidateComponents(components)

	return err
}
