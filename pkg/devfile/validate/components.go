package validate

import (
	"fmt"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/odo/util/pushtarget"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Errors
var (
	ErrorNoComponents              = "no components present"
	ErrorNoContainerComponent      = fmt.Sprintf("odo requires atleast one component of type '%s' in devfile", common.ContainerComponentType)
	ErrorDuplicateVolumeComponents = "duplicate volume components present in devfile"
	ErrorInvalidVolumeSize         = "size %s for volume component %s is invalid: %v. Example - 2Gi, 1024Mi"
)

// validateComponents validates all the devfile components
func validateComponents(components []common.DevfileComponent) error {

	// components cannot be empty
	if len(components) < 1 {
		return fmt.Errorf(ErrorNoComponents)
	}

	processedVolumes := make(map[string]bool)

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
				if !pushtarget.IsPushTargetDocker() && len(component.Volume.Size) > 0 {
					// Only validate on Kubernetes since Docker volumes do not use sizes
					// We use the Kube API for validation because there are so many ways to
					// express storage in Kubernetes. For reference, you may check doc
					// https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
					if _, err := resource.ParseQuantity(component.Volume.Size); err != nil {
						return fmt.Errorf(ErrorInvalidVolumeSize, component.Volume.Size, component.Volume.Name, err)
					}
				}
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
