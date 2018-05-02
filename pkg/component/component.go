package component

import (
	"fmt"
	"net/url"
	"strings"

	buildv1 "github.com/openshift/api/build/v1"
	"github.com/pkg/errors"
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/config"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/storage"
	urlpkg "github.com/redhat-developer/odo/pkg/url"
	"github.com/redhat-developer/odo/pkg/util"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

// componentSourceURLAnnotation is an source url from which component was build
// it can be also file://
const componentSourceURLAnnotation = "app.kubernetes.io/url"
const componentSourceTypeAnnotation = "app.kubernetes.io/component-source-type"

// ComponentInfo holds all important information about one component
type ComponentInfo struct {
	Name string
	Type string
}

func CreateFromGit(client *occlient.Client, name string, ctype string, url string, applicationName string) error {
	labels := componentlabels.GetLabels(name, applicationName, true)
	// save component type as label
	labels[componentlabels.ComponentTypeLabel] = ctype

	// save source path as annotation
	annotations := map[string]string{componentSourceURLAnnotation: url}
	annotations[componentSourceTypeAnnotation] = "git"

	err := client.NewAppS2I(name, ctype, url, labels, annotations)
	if err != nil {
		return errors.Wrapf(err, "unable to create git component %s", name)
	}
	return nil
}

// CreateFromPath create new component with source or binary from the given local path
// sourceType indicates the source type of the component and can be either local or binary
func CreateFromPath(client *occlient.Client, name string, ctype string, path string, applicationName string, sourceType string) error {
	labels := componentlabels.GetLabels(name, applicationName, true)
	// save component type as label
	labels[componentlabels.ComponentTypeLabel] = ctype

	// save source path as annotation
	sourceURL := url.URL{Scheme: "file", Path: path}
	annotations := map[string]string{componentSourceURLAnnotation: sourceURL.String()}
	annotations[componentSourceTypeAnnotation] = sourceType

	err := client.NewAppS2I(name, ctype, "", labels, annotations)
	if err != nil {
		return err
	}

	return nil

}

// Delete whole component
func Delete(client *occlient.Client, name string, applicationName string, projectName string) (string, error) {

	cfg, err := config.New()
	if err != nil {
		return "", errors.Wrapf(err, "unable to create new configuration to delete %s", name)
	}

	labels := componentlabels.GetLabels(name, applicationName, false)

	output, err := client.Delete("all", "", labels)
	if err != nil {
		return "", errors.Wrapf(err, "error deleting component %s", name)
	}

	// Get a list of all active components
	components, err := List(client, applicationName, projectName)
	if err != nil {
		return "", errors.Wrapf(err, "unable to retrieve list of components")
	}

	// We will *only* set a new component if either len(components) is zero, or the
	// current component matches the one being deleted.
	if current := cfg.GetActiveComponent(applicationName, projectName); current == name || len(components) == 0 {

		// If there's more than one component, set it to the first one..
		if len(components) > 0 {
			err = cfg.SetActiveComponent(components[0].Name, applicationName, projectName)

			if err != nil {
				return "", errors.Wrapf(err, "unable to set current component to '%s'", name)
			}
		} else {
			// Unset to blank
			err = cfg.UnsetActiveComponent(applicationName, projectName)
			if err != nil {
				return "", errors.Wrapf(err, "error unsetting current component while deleting %s", name)
			}

		}
	}

	return output, nil
}

func SetCurrent(client *occlient.Client, name string, applicationName string, projectName string) error {
	cfg, err := config.New()
	if err != nil {
		return errors.Wrapf(err, "unable to set current component %s", name)
	}

	err = cfg.SetActiveComponent(name, applicationName, projectName)
	if err != nil {
		return errors.Wrapf(err, "unable to set current component %s", name)
	}

	return nil
}

// GetCurrent component in active application
// returns "" if there is no active component
func GetCurrent(client *occlient.Client, applicationName string, projectName string) (string, error) {
	cfg, err := config.New()
	if err != nil {
		return "", errors.Wrap(err, "unable to get config")
	}
	currentComponent := cfg.GetActiveComponent(applicationName, projectName)

	return currentComponent, nil

}

// PushLocal start new build and push local dir as a source for local or a binary file for a binary build
// asFile indicates if it is a binary component or not
func PushLocal(client *occlient.Client, componentName string, dir string, asFile bool) error {
	err := client.StartBinaryBuild(componentName, dir, asFile)
	if err != nil {
		return errors.Wrap(err, "unable to start build")
	}
	return nil
}

// RebuildGit rebuild git component from the git repo that it was created with
func RebuildGit(client *occlient.Client, componentName string) error {
	if err := client.StartBuild(componentName); err != nil {
		return errors.Wrapf(err, "unable to rebuild %s", componentName)
	}
	return nil
}

// GetComponentType returns type of component in given application and project
func GetComponentType(client *occlient.Client, componentName string, applicationName string, projectName string) (string, error) {
	// filter according to component and application name
	selector := fmt.Sprintf("%s=%s,%s=%s", componentlabels.ComponentLabel, componentName, applabels.ApplicationLabel, applicationName)
	ctypes, err := client.GetLabelValues(projectName, componentlabels.ComponentTypeLabel, selector)
	if err != nil {
		return "", errors.Wrap(err, "unable to get type of %s component")
	}
	if len(ctypes) < 1 {
		// no type returned
		return "", errors.Wrap(err, "unable to find type of %s component")

	}
	// check if all types are the same
	// it should be as we are secting only exactly one component, and it doesn't make sense
	// to have one component labeled with different component type labels
	for _, ctype := range ctypes {
		if ctypes[0] != ctype {
			return "", errors.Wrap(err, "data mismatch: %s component has objects with different types")
		}

	}
	return ctypes[0], nil
}

// List lists components in active application
func List(client *occlient.Client, applicationName string, projectName string) ([]ComponentInfo, error) {
	applicationSelector := fmt.Sprintf("%s=%s", applabels.ApplicationLabel, applicationName)
	componentNames, err := client.GetLabelValues(projectName, componentlabels.ComponentLabel, applicationSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to list components")
	}

	components := []ComponentInfo{}

	for _, name := range componentNames {
		ctype, err := GetComponentType(client, name, applicationName, projectName)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to list components")
		}
		components = append(components, ComponentInfo{Name: name, Type: ctype})
	}

	return components, nil
}

// GetComponentSource what source type given component uses
// The first returned string is component source type ("git" or "local" or "binary")
// The second returned string is a source (url to git repository or local path or path to binary)
func GetComponentSource(client *occlient.Client, componentName string, applicationName string, projectName string) (string, string, error) {
	bc, err := client.GetBuildConfig(componentName, projectName)
	if err != nil {
		return "", "", errors.Wrapf(err, "unable to get source path for component %s", componentName)
	}

	var sourceType string
	var sourcePath string

	switch bc.Spec.Source.Type {
	case buildv1.BuildSourceGit:
		sourceType = "git"
		sourcePath = bc.Spec.Source.Git.URI
	case buildv1.BuildSourceBinary:
		sourceType = bc.ObjectMeta.Annotations[componentSourceTypeAnnotation]
		sourcePath = bc.ObjectMeta.Annotations[componentSourceURLAnnotation]
		if sourcePath == "" {
			return "", "", fmt.Errorf("unsupported BuildConfig.Spec.Source.Type %s", bc.Spec.Source.Type)
		}
	default:
		return "", "", fmt.Errorf("unsupported BuildConfig.Spec.Source.Type %s", bc.Spec.Source.Type)
	}

	log.Debugf("Component %s source type is %s (%s)", componentName, sourceType, sourcePath)
	return sourceType, sourcePath, nil
}

// Update updates the requested component
// Component name is the name component to be updated
// to indicates what type of source type the component source is changing to e.g from git to local or local to binary
// source indicates path of the source directory or binary or the git URL
func Update(client *occlient.Client, componentName string, to string, source string) error {
	var err error
	projectName := client.GetCurrentProjectName()
	// save source path as annotation
	var annotations map[string]string
	if to == "git" {
		annotations = map[string]string{componentSourceURLAnnotation: source}
		annotations[componentSourceTypeAnnotation] = to
		err = client.UpdateBuildConfig(componentName, projectName, source, annotations)
	} else if to == "local" {
		sourceURL := url.URL{Scheme: "file", Path: source}
		annotations = map[string]string{componentSourceURLAnnotation: sourceURL.String()}
		annotations[componentSourceTypeAnnotation] = to
		err = client.UpdateBuildConfig(componentName, projectName, "", annotations)
	} else if to == "binary" {
		sourceURL := url.URL{Scheme: "file", Path: source}
		annotations = map[string]string{componentSourceURLAnnotation: sourceURL.String()}
		annotations[componentSourceTypeAnnotation] = to
		err = client.UpdateBuildConfig(componentName, projectName, "", annotations)
	}
	if err != nil {
		return errors.Wrap(err, "unable to update the component")
	}
	return nil
}

// Checks whether a component with the given name exists in the current application or not
// componentName is the component name to perform check for
// The first returned parameter is a bool indicating if a component with the given name already exists or not
// The second returned parameter is the error that might occurs while execution
func Exists(client *occlient.Client, componentName, applicationName, projectName string) (bool, error) {
	componentList, err := List(client, applicationName, projectName)
	if err != nil {
		return false, errors.Wrap(err, "unable to get the component list")
	}
	for _, component := range componentList {
		if component.Name == componentName {
			return true, nil
		}
	}
	return false, nil
}

// LinkInfo contains the information about the link being created between
// components
type LinkInfo struct {
	SourceComponent string
	TargetComponent string
	Envs            []string
}

// Link injects connection information of the target component into the source
// component as environment variables
func Link(client *occlient.Client, sourceComponent, targetComponent, applicationName string) (*LinkInfo, error) {
	// Get Service associated with the target component
	serviceLabels := componentlabels.GetLabels(targetComponent, applicationName, false)
	serviceSelector := util.ConvertLabelsToSelector(serviceLabels)
	targetService, err := client.GetOneServiceFromSelector(serviceSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get service associated with the component %v", targetComponent)
	}

	// Generate environment variables to inject
	linkHostKey := fmt.Sprintf("COMPONENT_%v_HOST", strings.ToUpper(targetComponent))
	linkHostValue := targetService.Name

	linkPortKey := fmt.Sprintf("COMPONENT_%v_PORT", strings.ToUpper(targetComponent))
	linkPort, err := getPortFromService(targetService)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get port from Service %v", targetService.Name)
	}
	linkPortValue := fmt.Sprint(linkPort)

	// Inject environment variable to source component
	dcLabels := componentlabels.GetLabels(sourceComponent, applicationName, false)
	dcSelector := util.ConvertLabelsToSelector(dcLabels)
	sourceDC, err := client.GetOneDeploymentConfigFromSelector(dcSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get Deployment Config for component %v", sourceComponent)
	}

	if err := client.AddEnvironmentVariablesToDeploymentConfig(
		[]corev1.EnvVar{
			{
				Name:  linkHostKey,
				Value: linkHostValue,
			},
			{
				Name:  linkPortKey,
				Value: linkPortValue,
			},
		}, sourceDC); err != nil {
		return nil, errors.Wrapf(err, "unable to add environment variables to Deployment Config %v", sourceDC.Name)
	}

	return &LinkInfo{
		SourceComponent: sourceComponent,
		TargetComponent: targetComponent,
		Envs: []string{
			fmt.Sprintf("%v=%v", linkHostKey, linkHostValue),
			fmt.Sprintf("%v=%v", linkPortKey, linkPortValue),
		},
	}, nil
}

// getPortFromService returns the first port listed in the service
func getPortFromService(service *corev1.Service) (int32, error) {
	numServicePorts := len(service.Spec.Ports)
	if numServicePorts != 1 {
		return 0, fmt.Errorf("expected exactly one port in the service %v, but got %v", service.Name, numServicePorts)
	}

	return service.Spec.Ports[0].Port, nil

}

// Get Component Description
func GetComponentDesc(client *occlient.Client, currentComponent string, currentApplication string, currentProject string) (componentType string, path string, componentURL string, appStore []storage.StorageInfo, err error) {
	// Component Type
	componentType, err = GetComponentType(client, currentComponent, currentApplication, currentProject)
	if err != nil {
		return "", "", "", nil, errors.Wrap(err, "unable to get source path")
	}
	// Source
	_, path, err = GetComponentSource(client, currentComponent, currentApplication, currentProject)
	if err != nil {
		return "", "", "", nil, errors.Wrap(err, "unable to get source path")
	}
	// URL
	urlList, err := urlpkg.List(client, currentComponent, currentApplication)
	if len(urlList) != 0 {
		componentURL = urlList[0].URL
	}
	if err != nil {
		return "", "", "", nil, errors.Wrap(err, "unable to get url list")
	}
	//Storage
	appStore, err = storage.List(client, currentApplication, currentComponent)
	if err != nil {
		return "", "", "", nil, errors.Wrap(err, "unable to get storage list")
	}

	return componentType, path, componentURL, appStore, nil
}
