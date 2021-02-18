package validation

import (
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	// EnvProjectsSrc is the env defined for path to the project source in a component container
	EnvProjectsSrc = "PROJECT_SOURCE"

	// EnvProjectsRoot is the env defined for project mount in a component container when component's mountSources=true
	EnvProjectsRoot = "PROJECTS_ROOT"
)

// ValidateComponents validates that the components
// 1. makes sure the container components reference a valid volume component if it uses volume mounts
// 2. makes sure the volume components are unique
// 3. checks the URI specified in openshift components and kubernetes components are with valid format
// 4. makes sure the component name is unique
func ValidateComponents(components []v1alpha2.Component) error {

	processedVolumes := make(map[string]bool)
	processedVolumeMounts := make(map[string][]string)
	processedEndPointName := make(map[string]bool)
	processedEndPointPort := make(map[int]bool)

	err := v1alpha2.CheckDuplicateKeys(components)
	if err != nil {
		return err
	}

	for _, component := range components {
		switch {
		case component.Container != nil:
			// Process all the volume mounts in container components to validate them later
			for _, volumeMount := range component.Container.VolumeMounts {
				processedVolumeMounts[component.Name] = append(processedVolumeMounts[component.Name], volumeMount.Name)

			}

			// Check if any containers are customizing the reserved PROJECT_SOURCE or PROJECTS_ROOT env
			for _, env := range component.Container.Env {
				if env.Name == EnvProjectsSrc {
					return &ReservedEnvError{envName: EnvProjectsSrc, componentName: component.Name}
				} else if env.Name == EnvProjectsRoot {
					return &ReservedEnvError{envName: EnvProjectsRoot, componentName: component.Name}
				}
			}

			err := validateEndpoints(component.Container.Endpoints, processedEndPointPort, processedEndPointName)
			if err != nil {
				return err
			}
		case component.Volume != nil:
			processedVolumes[component.Name] = true
			if len(component.Volume.Size) > 0 {
				// We use the Kube API for validation because there are so many ways to
				// express storage in Kubernetes. For reference, you may check doc
				// https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
				if _, err := resource.ParseQuantity(component.Volume.Size); err != nil {
					return &InvalidVolumeError{name: component.Name, reason: fmt.Sprintf("size %s for volume component is invalid, %v. Example - 2Gi, 1024Mi", component.Volume.Size, err)}
				}
			}
		case component.Openshift != nil:
			if component.Openshift.Uri != "" {
				err := ValidateURI(component.Openshift.Uri)
				if err != nil {
					return err
				}
			}

			err := validateEndpoints(component.Openshift.Endpoints, processedEndPointPort, processedEndPointName)
			if err != nil {
				return err
			}
		case component.Kubernetes != nil:
			if component.Kubernetes.Uri != "" {
				err := ValidateURI(component.Kubernetes.Uri)
				if err != nil {
					return err
				}
			}
			err := validateEndpoints(component.Kubernetes.Endpoints, processedEndPointPort, processedEndPointName)
			if err != nil {
				return err
			}
		case component.Plugin != nil:
			if component.Plugin.RegistryUrl != "" {
				err := ValidateURI(component.Plugin.RegistryUrl)
				if err != nil {
					return err
				}
			}
		}

	}

	// Check if the volume mounts mentioned in the containers are referenced by a volume component
	var invalidVolumeMountsErr string
	for componentName, volumeMountNames := range processedVolumeMounts {
		for _, volumeMountName := range volumeMountNames {
			if !processedVolumes[volumeMountName] {
				invalidVolumeMountsErr += fmt.Sprintf("\nvolume mount %s belonging to the container component %s", volumeMountName, componentName)
			}
		}
	}

	if len(invalidVolumeMountsErr) > 0 {
		return &MissingVolumeMountError{errMsg: invalidVolumeMountsErr}
	}

	return nil
}
