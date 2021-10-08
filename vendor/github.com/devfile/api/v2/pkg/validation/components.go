package validation

import (
	"fmt"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/hashicorp/go-multierror"
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
// 5. makes sure the image dockerfile component git src has at most one remote
func ValidateComponents(components []v1alpha2.Component) (returnedErr error) {

	processedVolumes := make(map[string]bool)
	processedVolumeMounts := make(map[string][]string)
	processedEndPointName := make(map[string]bool)
	processedEndPointPort := make(map[int]bool)
	processedComponentWithVolumeMounts := make(map[string]v1alpha2.Component)

	err := v1alpha2.CheckDuplicateKeys(components)
	if err != nil {
		returnedErr = multierror.Append(returnedErr, err)
	}

	for _, component := range components {
		switch {
		case component.Container != nil:
			// Process all the volume mounts in container components to validate them later
			for _, volumeMount := range component.Container.VolumeMounts {
				processedVolumeMounts[component.Name] = append(processedVolumeMounts[component.Name], volumeMount.Name)
				processedComponentWithVolumeMounts[component.Name] = component

			}

			// Check if any containers are customizing the reserved PROJECT_SOURCE or PROJECTS_ROOT env
			for _, env := range component.Container.Env {
				if env.Name == EnvProjectsSrc {
					reservedEnvErr := &ReservedEnvError{envName: EnvProjectsSrc, componentName: component.Name}
					returnedErr = multierror.Append(returnedErr, reservedEnvErr)
				} else if env.Name == EnvProjectsRoot {
					reservedEnvErr := &ReservedEnvError{envName: EnvProjectsRoot, componentName: component.Name}
					returnedErr = multierror.Append(returnedErr, reservedEnvErr)
				}
			}

			err := validateEndpoints(component.Container.Endpoints, processedEndPointPort, processedEndPointName)
			if len(err) > 0 {
				for _, endpointErr := range err {
					returnedErr = multierror.Append(returnedErr, resolveErrorMessageWithImportAttributes(endpointErr, component.Attributes))
				}
			}
		case component.Volume != nil:
			processedVolumes[component.Name] = true
			if len(component.Volume.Size) > 0 {
				// We use the Kube API for validation because there are so many ways to
				// express storage in Kubernetes. For reference, you may check doc
				// https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
				if _, err := resource.ParseQuantity(component.Volume.Size); err != nil {
					invalidVolErr := &InvalidVolumeError{name: component.Name, reason: fmt.Sprintf("size %s for volume component is invalid, %v. Example - 2Gi, 1024Mi", component.Volume.Size, err)}
					returnedErr = multierror.Append(returnedErr, resolveErrorMessageWithImportAttributes(invalidVolErr, component.Attributes))
				}
			}
		case component.Openshift != nil:
			if component.Openshift.Uri != "" {
				err := ValidateURI(component.Openshift.Uri)
				if err != nil {
					returnedErr = multierror.Append(returnedErr, resolveErrorMessageWithImportAttributes(err, component.Attributes))
				}
			}

			err := validateEndpoints(component.Openshift.Endpoints, processedEndPointPort, processedEndPointName)
			if len(err) > 0 {
				for _, endpointErr := range err {
					returnedErr = multierror.Append(returnedErr, resolveErrorMessageWithImportAttributes(endpointErr, component.Attributes))
				}
			}
		case component.Kubernetes != nil:
			if component.Kubernetes.Uri != "" {
				err := ValidateURI(component.Kubernetes.Uri)
				if err != nil {
					returnedErr = multierror.Append(returnedErr, resolveErrorMessageWithImportAttributes(err, component.Attributes))
				}
			}
			err := validateEndpoints(component.Kubernetes.Endpoints, processedEndPointPort, processedEndPointName)
			if len(err) > 0 {
				for _, endpointErr := range err {
					returnedErr = multierror.Append(returnedErr, resolveErrorMessageWithImportAttributes(endpointErr, component.Attributes))
				}
			}
		case component.Image != nil:
			var gitSource v1alpha2.GitLikeProjectSource
			if component.Image.Dockerfile != nil && component.Image.Dockerfile.Git != nil {
				gitSource = component.Image.Dockerfile.Git.GitLikeProjectSource
				if err := validateSingleRemoteGitSrc("component", component.Name, gitSource); err != nil {
					returnedErr = multierror.Append(returnedErr, resolveErrorMessageWithImportAttributes(err, component.Attributes))
				}
			}
		case component.Plugin != nil:
			if component.Plugin.RegistryUrl != "" {
				err := ValidateURI(component.Plugin.RegistryUrl)
				if err != nil {
					returnedErr = multierror.Append(returnedErr, resolveErrorMessageWithImportAttributes(err, component.Attributes))
				}
			}
		}

	}

	// Check if the volume mounts mentioned in the containers are referenced by a volume component
	var invalidVolumeMountsErrList []string
	for componentName, volumeMountNames := range processedVolumeMounts {
		for _, volumeMountName := range volumeMountNames {
			if !processedVolumes[volumeMountName] {
				missingVolumeMountErr := fmt.Errorf("volume mount %s belonging to the container component %s", volumeMountName, componentName)
				newErr := resolveErrorMessageWithImportAttributes(missingVolumeMountErr, processedComponentWithVolumeMounts[componentName].Attributes)
				invalidVolumeMountsErrList = append(invalidVolumeMountsErrList, newErr.Error())
			}
		}
	}

	if len(invalidVolumeMountsErrList) > 0 {
		invalidVolumeMountsErr := fmt.Sprintf("\n%s", strings.Join(invalidVolumeMountsErrList, "\n"))
		returnedErr = multierror.Append(returnedErr, &MissingVolumeMountError{errMsg: invalidVolumeMountsErr})
	}

	return returnedErr
}
