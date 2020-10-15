package validate

import (
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// validateComponents validates the devfile components:
// 1. there should be at least one component
// 2. there should be at least one container component
func validateComponents(components []common.DevfileComponent) error {

	// components cannot be empty
	if len(components) < 1 {
		return &NoComponentsError{}
	}

	// Check if component of type container is present
	for _, component := range components {
		if component.Container != nil {
			return nil
		}
	}

	return &NoContainerComponentError{}
}
