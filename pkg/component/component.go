package component

import (
	"fmt"
	"net/url"

	buildv1 "github.com/openshift/api/build/v1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/config"
	"github.com/redhat-developer/odo/pkg/occlient"
)

// componentSourceURLAnnotation is an source url from which component was build
// it can be also file://
const componentSourceURLAnnotation = "app.kubernetes.io/url"

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

	err := client.NewAppS2I(name, ctype, url, labels, annotations)
	if err != nil {
		return errors.Wrapf(err, "unable to create git component %s", name)
	}
	return nil
}

// CreateFromDir create new component with source from local directory
func CreateFromDir(client *occlient.Client, name string, ctype string, dir string, applicationName string) error {
	labels := componentlabels.GetLabels(name, applicationName, true)
	// save component type as label
	labels[componentlabels.ComponentTypeLabel] = ctype

	// save source path as annotation
	sourceURL := url.URL{Scheme: "file", Path: dir}
	annotations := map[string]string{componentSourceURLAnnotation: sourceURL.String()}

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

// PushLocal start new build and push local dir as a source for build
func PushLocal(client *occlient.Client, componentName string, dir string) error {
	err := client.StartBinaryBuild(componentName, dir)
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
// The first returned string is component source type ("git" or "local")
// The second returned string is a source (url to git repository or local path)
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
		sourceType = "local"
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
// to indicates what type of source type the component source is changing to e.g from git to local
// source indicates dir or the git URL
func Update(client *occlient.Client, componentName string, to string, source string) error {
	var err error
	projectName := client.GetCurrentProjectName()
	// save source path as annotation
	var annotations map[string]string
	if to == "git" {
		annotations = map[string]string{componentSourceURLAnnotation: source}
		err = client.UpdateBuildConfig(componentName, projectName, source, annotations)
	} else if to == "dir" {
		sourceURL := url.URL{Scheme: "file", Path: source}
		annotations = map[string]string{componentSourceURLAnnotation: sourceURL.String()}
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
