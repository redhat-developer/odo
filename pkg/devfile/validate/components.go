package validate

import (
	"fmt"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// Errors
var (
	ErrorNoComponents           = "no components present"
	ErrorNoDockerImageComponent = fmt.Sprintf("odo requires atleast one component of type '%s' in devfile", common.DevfileComponentTypeDockerimage)
)

// ValidateComponents validates all the devfile components
func ValidateComponents(components []common.DevfileComponent) error {

	// components cannot be empty
	if len(components) < 1 {
		return fmt.Errorf(ErrorNoComponents)
	}

	// Check wether component of type dockerimage is present
	isDockerImageComponentPresent := false
	for _, component := range components {
		if component.Type == common.DevfileComponentTypeDockerimage {
			isDockerImageComponentPresent = true
			break
		}
	}

	if !isDockerImageComponentPresent {
		return fmt.Errorf(ErrorNoDockerImageComponent)
	}

	// Successful
	return nil
}
