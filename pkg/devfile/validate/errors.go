package validate

import (
	"fmt"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// NoComponentsError returns an error if no component is found
type NoComponentsError struct {
}

func (e *NoComponentsError) Error() string {
	return "no components present"
}

// NoContainerComponentError returns an error if no container component is found
type NoContainerComponentError struct {
}

func (e *NoContainerComponentError) Error() string {
	return fmt.Sprintf("odo requires atleast one component of type '%s' in devfile", common.ContainerComponentType)
}

// DuplicateVolumeComponentsError returns an error if duplicate volume components are found
type DuplicateVolumeComponentsError struct {
}

func (e *DuplicateVolumeComponentsError) Error() string {
	return "duplicate volume components present in devfile"
}

// InvalidVolumeSizeError returns an error if volume component has an invalid size
type InvalidVolumeSizeError struct {
	size            string
	componentName   string
	validationError error
}

func (e *InvalidVolumeSizeError) Error() string {
	return fmt.Sprintf("size %s for volume component %s is invalid: %v. Example - 2Gi, 1024Mi", e.size, e.componentName, e.validationError)
}

// MissingVolumeMountError returns an error if the container volume mount does not reference a valid volume component
type MissingVolumeMountError struct {
	volumeName string
}

func (e *MissingVolumeMountError) Error() string {
	return fmt.Sprintf("unable to find volume mount %s in devfile volume components", e.volumeName)
}

// InvalidEventError returns an error if the devfile event type has invalid events
type InvalidEventError struct {
	eventType string
	errorMsg  string
}

func (e *InvalidEventError) Error() string {
	return fmt.Sprintf("%s type events is invalid: %s", e.eventType, e.errorMsg)
}
