package validate

import (
	"strings"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/odo/util/pushtarget"
	"k8s.io/apimachinery/pkg/api/resource"
)

// validateComponents validates all the devfile components
func validateComponents(components []common.DevfileComponent) error {

	// components cannot be empty
	if len(components) < 1 {
		return &NoComponentsError{}
	}

	processedVolumes := make(map[string]bool)
	processedVolumeMounts := make(map[string]bool)

	// Check if component of type container is present
	// and if volume components are unique
	isContainerComponentPresent := false
	for _, component := range components {
		if component.Container != nil {
			isContainerComponentPresent = true

			for _, volumeMount := range component.Container.VolumeMounts {
				if _, ok := processedVolumeMounts[volumeMount.Name]; !ok {
					processedVolumeMounts[volumeMount.Name] = true
				}
			}
		}

		if component.Volume != nil {
			if _, ok := processedVolumes[component.Name]; !ok {
				processedVolumes[component.Name] = true
				if !pushtarget.IsPushTargetDocker() && len(component.Volume.Size) > 0 {
					// Only validate on Kubernetes since Docker volumes do not use sizes
					// We use the Kube API for validation because there are so many ways to
					// express storage in Kubernetes. For reference, you may check doc
					// https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
					if _, err := resource.ParseQuantity(component.Volume.Size); err != nil {
						return &InvalidVolumeSizeError{size: component.Volume.Size, componentName: component.Name, validationError: err}
					}
				}
			} else {
				return &DuplicateVolumeComponentsError{}
			}
		}
	}

	if !isContainerComponentPresent {
		return &NoContainerComponentError{}
	}

	var invalidVolumeMounts []string
	for volumeMountName := range processedVolumeMounts {
		if _, ok := processedVolumes[volumeMountName]; !ok {
			invalidVolumeMounts = append(invalidVolumeMounts, volumeMountName)
		}
	}

	if len(invalidVolumeMounts) > 0 {
		return &MissingVolumeMountError{volumeName: strings.Join(invalidVolumeMounts, ",")}
	}

	// Successful
	return nil
}
