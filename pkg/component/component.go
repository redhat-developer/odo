package component

import (
	"fmt"
	"net/url"

	buildv1 "github.com/openshift/api/build/v1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/redhat-developer/ocdev/pkg/application"
	"github.com/redhat-developer/ocdev/pkg/config"
	"github.com/redhat-developer/ocdev/pkg/occlient"
	"github.com/redhat-developer/ocdev/pkg/project"
)

// ComponentLabel is a label key used to identify component
const ComponentLabel = "app.kubernetes.io/component-name"

// componentTypeLabel is kubernetes that identifies type of a component
const componentTypeLabel = "app.kubernetes.io/component-type"

// componentSourceURLAnnotation is an source url from which component was build
// it can be also file://
const componentSourceURLAnnotation = "app.kubernetes.io/url"

// ComponentInfo holds all important information about one component
type ComponentInfo struct {
	Name string
	Type string
}

// GetLabels return labels that should be applied to every object for given component in active application
// additional labels are used only for creating object
// if you are creating something use additional=true
// if you need labels to filter component that use additional=false
func GetLabels(componentName string, applicationName string, additional bool) map[string]string {
	labels := application.GetLabels(applicationName, additional)
	labels[ComponentLabel] = componentName
	return labels
}

func CreateFromGit(client *occlient.Client, name string, ctype string, url string) error {
	// if current application doesn't exist, create it
	// this can happen when ocdev is started form clean state
	// and default application is used (first command run is ocdev create)
	currentApplication, err := application.GetCurrentOrDefault(client)
	if err != nil {
		return errors.Wrapf(err, "unable to create git component %s", name)
	}
	exists, err := application.Exists(client, currentApplication)
	if err != nil {
		return errors.Wrapf(err, "unable to create git component %s", name)
	}
	if !exists {
		err = application.Create(client, currentApplication)
		if err != nil {
			return errors.Wrapf(err, "unable to create git component %s", name)
		}
	}

	labels := GetLabels(name, currentApplication, true)
	// save component type as label
	labels[componentTypeLabel] = ctype

	// save source path as annotation
	annotations := map[string]string{componentSourceURLAnnotation: url}

	err = client.NewAppS2I(name, ctype, url, labels, annotations)
	if err != nil {
		return errors.Wrapf(err, "unable to create git component %s", name)
	}

	fmt.Println("please wait, building component...")

	//get the latest build name for following
	buildName, err := client.GetLatestBuildName(name)
	if err != nil {
		return errors.Wrap(err, "unable to follow build logs")
	}
	err = client.FollowBuildLog(buildName)
	if err != nil {
		return errors.Wrap(err, "unable to follow build logs")
	}

	return nil
}

// CreateFromDir create new component with source from local directory
func CreateFromDir(client *occlient.Client, name string, ctype string, dir string) error {
	currentApplication, err := application.GetCurrentOrDefault(client)
	if err != nil {
		return errors.Wrapf(err, "unable to create component %s from local path", name, dir)
	}

	exists, err := application.Exists(client, currentApplication)
	if err != nil {
		return errors.Wrapf(err, "unable to create component %s from local path", name, dir)
	}
	if !exists {
		err = application.Create(client, currentApplication)
		if err != nil {
			return errors.Wrapf(err, "unable to create component %s from local path", name, dir)
		}
	}

	labels := GetLabels(name, currentApplication, true)
	// save component type as label
	labels[componentTypeLabel] = ctype

	// save source path as annotation
	sourceURL := url.URL{Scheme: "file", Path: dir}
	annotations := map[string]string{componentSourceURLAnnotation: sourceURL.String()}

	err = client.NewAppS2I(name, ctype, "", labels, annotations)
	if err != nil {
		return err
	}

	fmt.Println("please wait, building component...")

	err = client.StartBinaryBuild(name, dir)
	if err != nil {
		return err
	}
	return nil

}

// Delete whole component
func Delete(client *occlient.Client, name string) (string, error) {
	currentApplication, err := application.GetCurrentOrDefault(client)
	if err != nil {
		return "", errors.Wrapf(err, "unable to delete component %s", name)
	}

	currentProject := project.GetCurrent(client)

	cfg, err := config.New()
	if err != nil {
		return "", errors.Wrapf(err, "unable to delete component %s", name)
	}

	labels := GetLabels(name, currentApplication, false)

	output, err := client.Delete("all", "", labels)
	if err != nil {
		return "", errors.Wrapf(err, "unable to delete component %s", name)
	}

	err = cfg.SetActiveComponent("", currentProject, currentApplication)
	if err != nil {
		return "", errors.Wrapf(err, "unable to delete component %s", name)
	}

	return output, nil
}

func SetCurrent(client *occlient.Client, name string) error {
	cfg, err := config.New()
	if err != nil {
		return errors.Wrapf(err, "unable to set current component %s", name)
	}

	currentProject := project.GetCurrent(client)

	currentApplication, err := application.GetCurrent(client)
	if err != nil {
		return errors.Wrapf(err, "unable to set current component %s", name)
	}

	err = cfg.SetActiveComponent(name, currentApplication, currentProject)
	if err != nil {
		return errors.Wrapf(err, "unable to set current component %s", name)
	}

	return nil
}

// GetCurrent component in active application
// returns "" if there is no active component
func GetCurrent(client *occlient.Client) (string, error) {
	cfg, err := config.New()
	if err != nil {
		return "", errors.Wrap(err, "unable to get config")
	}
	currentApplication, err := application.GetCurrentOrDefault(client)
	if err != nil {
		return "", errors.Wrap(err, "unable to get active application")
	}

	currentProject := project.GetCurrent(client)
	currentComponent := cfg.GetActiveComponent(currentApplication, currentProject)

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
	selector := fmt.Sprintf("%s=%s,%s=%s", ComponentLabel, componentName, application.ApplicationLabel, applicationName)
	ctypes, err := client.GetLabelValues(projectName, componentTypeLabel, selector)
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
func List(client *occlient.Client) ([]ComponentInfo, error) {
	currentProject := project.GetCurrent(client)

	currentApplication, err := application.GetCurrent(client)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list components")
	}

	applicationSelector := fmt.Sprintf("%s=%s", application.ApplicationLabel, currentApplication)
	componentNames, err := client.GetLabelValues(currentProject, ComponentLabel, applicationSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to list components")
	}

	components := []ComponentInfo{}

	for _, name := range componentNames {
		ctype, err := GetComponentType(client, name, currentApplication, currentProject)
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
