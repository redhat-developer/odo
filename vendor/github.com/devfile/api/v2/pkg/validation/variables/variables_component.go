package variables

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// ValidateAndReplaceForComponents validates the components data for global variable references and replaces them with the variable value
// Returns a map of component names and invalid variable references if present.
func ValidateAndReplaceForComponents(variables map[string]string, components []v1alpha2.Component) map[string][]string {

	componentsWarningMap := make(map[string][]string)

	for i := range components {
		var err error

		// Validate various component types
		switch {
		case components[i].Container != nil:
			if err = validateAndReplaceForContainerComponent(variables, components[i].Container); err != nil {
				if verr, ok := err.(*InvalidKeysError); ok {
					componentsWarningMap[components[i].Name] = verr.Keys
				}
			}
		case components[i].Kubernetes != nil:
			if err = validateAndReplaceForKubernetesComponent(variables, components[i].Kubernetes); err != nil {
				if verr, ok := err.(*InvalidKeysError); ok {
					componentsWarningMap[components[i].Name] = verr.Keys
				}
			}
		case components[i].Openshift != nil:
			if err = validateAndReplaceForOpenShiftComponent(variables, components[i].Openshift); err != nil {
				if verr, ok := err.(*InvalidKeysError); ok {
					componentsWarningMap[components[i].Name] = verr.Keys
				}
			}
		case components[i].Volume != nil:
			if err = validateAndReplaceForVolumeComponent(variables, components[i].Volume); err != nil {
				if verr, ok := err.(*InvalidKeysError); ok {
					componentsWarningMap[components[i].Name] = verr.Keys
				}
			}
		}
	}

	return componentsWarningMap
}

// validateAndReplaceForContainerComponent validates the container component data for global variable references and replaces them with the variable value
func validateAndReplaceForContainerComponent(variables map[string]string, container *v1alpha2.ContainerComponent) error {
	var err error

	invalidKeys := make(map[string]bool)

	if container != nil {
		// Validate container image
		if container.Image, err = validateAndReplaceDataWithVariable(container.Image, variables); err != nil {
			checkForInvalidError(invalidKeys, err)
		}

		// Validate container commands
		for i := range container.Command {
			if container.Command[i], err = validateAndReplaceDataWithVariable(container.Command[i], variables); err != nil {
				checkForInvalidError(invalidKeys, err)
			}
		}

		// Validate container args
		for i := range container.Args {
			if container.Args[i], err = validateAndReplaceDataWithVariable(container.Args[i], variables); err != nil {
				checkForInvalidError(invalidKeys, err)
			}
		}

		// Validate memory limit
		if container.MemoryLimit, err = validateAndReplaceDataWithVariable(container.MemoryLimit, variables); err != nil {
			checkForInvalidError(invalidKeys, err)
		}

		// Validate memory request
		if container.MemoryRequest, err = validateAndReplaceDataWithVariable(container.MemoryRequest, variables); err != nil {
			checkForInvalidError(invalidKeys, err)
		}

		// Validate source mapping
		if container.SourceMapping, err = validateAndReplaceDataWithVariable(container.SourceMapping, variables); err != nil {
			checkForInvalidError(invalidKeys, err)
		}

		// Validate container env
		if len(container.Env) > 0 {
			if err = validateAndReplaceForEnv(variables, container.Env); err != nil {
				checkForInvalidError(invalidKeys, err)
			}
		}

		// Validate container volume mounts
		for i := range container.VolumeMounts {
			if container.VolumeMounts[i].Path, err = validateAndReplaceDataWithVariable(container.VolumeMounts[i].Path, variables); err != nil {
				checkForInvalidError(invalidKeys, err)
			}
		}

		// Validate container endpoints
		if len(container.Endpoints) > 0 {
			if err = validateAndReplaceForEndpoint(variables, container.Endpoints); err != nil {
				checkForInvalidError(invalidKeys, err)
			}
		}
	}

	return newInvalidKeysError(invalidKeys)
}

// validateAndReplaceForEnv validates the env data for global variable references and replaces them with the variable value
func validateAndReplaceForEnv(variables map[string]string, env []v1alpha2.EnvVar) error {

	invalidKeys := make(map[string]bool)

	for i := range env {
		var err error

		// Validate env name
		if env[i].Name, err = validateAndReplaceDataWithVariable(env[i].Name, variables); err != nil {
			checkForInvalidError(invalidKeys, err)
		}

		// Validate env value
		if env[i].Value, err = validateAndReplaceDataWithVariable(env[i].Value, variables); err != nil {
			checkForInvalidError(invalidKeys, err)
		}
	}

	return newInvalidKeysError(invalidKeys)
}

// validateAndReplaceForKubernetesComponent validates the kubernetes component data for global variable references and replaces them with the variable value
func validateAndReplaceForKubernetesComponent(variables map[string]string, kubernetes *v1alpha2.KubernetesComponent) error {
	var err error

	invalidKeys := make(map[string]bool)

	if kubernetes != nil {
		// Validate kubernetes uri
		if kubernetes.Uri, err = validateAndReplaceDataWithVariable(kubernetes.Uri, variables); err != nil {
			checkForInvalidError(invalidKeys, err)
		}

		// Validate kubernetes inlined
		if kubernetes.Inlined, err = validateAndReplaceDataWithVariable(kubernetes.Inlined, variables); err != nil {
			checkForInvalidError(invalidKeys, err)
		}

		// Validate kubernetes endpoints
		if len(kubernetes.Endpoints) > 0 {
			if err = validateAndReplaceForEndpoint(variables, kubernetes.Endpoints); err != nil {
				checkForInvalidError(invalidKeys, err)
			}
		}
	}

	return newInvalidKeysError(invalidKeys)
}

// validateAndReplaceForOpenShiftComponent validates the openshift component data for global variable references and replaces them with the variable value
func validateAndReplaceForOpenShiftComponent(variables map[string]string, openshift *v1alpha2.OpenshiftComponent) error {
	var err error

	invalidKeys := make(map[string]bool)

	if openshift != nil {
		// Validate openshift uri
		if openshift.Uri, err = validateAndReplaceDataWithVariable(openshift.Uri, variables); err != nil {
			checkForInvalidError(invalidKeys, err)
		}

		// Validate openshift inlined
		if openshift.Inlined, err = validateAndReplaceDataWithVariable(openshift.Inlined, variables); err != nil {
			checkForInvalidError(invalidKeys, err)
		}

		// Validate openshift endpoints
		if len(openshift.Endpoints) > 0 {
			if err = validateAndReplaceForEndpoint(variables, openshift.Endpoints); err != nil {
				checkForInvalidError(invalidKeys, err)
			}
		}
	}

	return newInvalidKeysError(invalidKeys)
}

// validateAndReplaceForVolumeComponent validates the volume component data for global variable references and replaces them with the variable value
func validateAndReplaceForVolumeComponent(variables map[string]string, volume *v1alpha2.VolumeComponent) error {
	var err error

	invalidKeys := make(map[string]bool)

	if volume != nil {
		// Validate volume size
		if volume.Size, err = validateAndReplaceDataWithVariable(volume.Size, variables); err != nil {
			checkForInvalidError(invalidKeys, err)
		}
	}

	return newInvalidKeysError(invalidKeys)
}
