package component

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/redhat-developer/ocdev/pkg/application"
	"github.com/redhat-developer/ocdev/pkg/config"
	"github.com/redhat-developer/ocdev/pkg/occlient"
)

// componentLabel is a label key used to identify component
const componentLabel = "app.kubernetes.io/component-name"

// componentTypeLabel is kubernetes that identifies type of a component
const componentTypeLabel = "app.kubernetes.io/component-type"

// ComponentInfo holds all important information about one component
type ComponentInfo struct {
	Name string
	Type string
}

// GetLabels return labels that should be applied to every object for given component in active application
// additional labels are used only for creating object
// if you are creating something use additional=true
// if you need labels to filter component that use additional=false
func GetLabels(componentName string, applicationName string, additional bool) (map[string]string, error) {
	labels, err := application.GetLabels(applicationName, additional)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get get labels for  component %s", componentName)
	}
	labels[componentLabel] = componentName

	return labels, nil
}

func CreateFromGit(name string, ctype string, url string) (string, error) {
	// if current application doesn't exist, create it
	// this can happen when ocdev is started form clean state
	// and default application is used (first command run is ocdev create)
	currentApplication, err := application.GetCurrentOrDefault()
	if err != nil {
		return "", errors.Wrapf(err, "unable to create git component %s", name)
	}
	exists, err := application.Exists(currentApplication)
	if err != nil {
		return "", errors.Wrapf(err, "unable to create git component %s", name)
	}
	if !exists {
		err = application.Create(currentApplication)
		if err != nil {
			return "", errors.Wrapf(err, "unable to create git component %s", name)
		}
	}

	labels, err := GetLabels(name, currentApplication, true)
	if err != nil {
		return "", errors.Wrapf(err, "unable to create git component %s", name)
	}

	// save component type as label
	labels[componentTypeLabel] = ctype

	output, err := occlient.NewAppS2I(name, ctype, url, labels)
	if err != nil {
		return "", errors.Wrapf(err, "unable to create git component %s", name)
	}

	if err = SetCurrent(name); err != nil {
		return "", errors.Wrapf(err, "unable to create git component %s", name)
	}
	return output, nil
}

func CreateEmpty(name string, ctype string) (string, error) {
	// if current application doesn't exist, create it
	// this can happen when ocdev is started form clean state
	// and default application is used (first command run is ocdev create)
	currentApplication, err := application.GetCurrentOrDefault()
	if err != nil {
		return "", errors.Wrapf(err, "unable to create git component %s", name)
	}
	exists, err := application.Exists(currentApplication)
	if err != nil {
		return "", errors.Wrapf(err, "unable to create git component %s", name)
	}
	if !exists {
		err = application.Create(currentApplication)
		if err != nil {
			return "", errors.Wrapf(err, "unable to create git component %s", name)
		}
	}

	labels, err := GetLabels(name, currentApplication, true)
	if err != nil {
		return "", errors.Wrapf(err, "unable to activate component %s created from git", name)
	}

	// save component type as label
	labels[componentTypeLabel] = ctype

	output, err := occlient.NewAppS2IEmpty(name, ctype, labels)
	if err != nil {
		return "", err
	}
	if err = SetCurrent(name); err != nil {
		return "", errors.Wrapf(err, "unable to activate empty component %s", name)
	}

	return output, nil
}

func CreateFromDir(name string, ctype, dir string) (string, error) {
	output, err := CreateEmpty(name, ctype)
	if err != nil {
		return "", errors.Wrap(err, "unable to get create empty component")
	}

	// TODO: it might not be ideal to print to stdout here
	fmt.Println(output)
	fmt.Println("please wait, building application...")

	output, err = occlient.StartBuild(name, dir)
	if err != nil {
		return "", err
	}
	fmt.Println(output)

	return "", nil

}

// Delete whole component
func Delete(name string) (string, error) {
	currentApplication, err := application.GetCurrentOrDefault()
	if err != nil {
		return "", errors.Wrapf(err, "unable to delete component %s", name)
	}

	currentProject, err := occlient.GetCurrentProjectName()
	if err != nil {
		return "", errors.Wrapf(err, "unable to delete component %s", name)
	}

	cfg, err := config.New()
	if err != nil {
		return "", errors.Wrapf(err, "unable to delete component %s", name)
	}

	labels, err := GetLabels(name, currentApplication, false)
	if err != nil {
		return "", errors.Wrapf(err, "unable to delete component %s", name)
	}

	output, err := occlient.Delete("all", "", labels)
	if err != nil {
		return "", errors.Wrapf(err, "unable to delete component %s", name)
	}

	err = cfg.SetActiveComponent("", currentProject, currentApplication)
	if err != nil {
		return "", errors.Wrapf(err, "unable to delete component %s", name)
	}

	return output, nil
}

func SetCurrent(name string) error {
	cfg, err := config.New()
	if err != nil {
		return errors.Wrapf(err, "unable to set current component %s", name)
	}

	currentProject, err := occlient.GetCurrentProjectName()
	if err != nil {
		return errors.Wrapf(err, "unable to set current component %s", name)
	}

	currentApplication, err := application.GetCurrent()
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
func GetCurrent() (string, error) {
	cfg, err := config.New()
	if err != nil {
		return "", errors.Wrap(err, "unable to get config")
	}
	currentApplication, err := application.GetCurrent()
	if err != nil {
		return "", errors.Wrap(err, "unable to get active application")
	}

	currentProject, err := occlient.GetCurrentProjectName()
	if err != nil {
		return "", errors.Wrap(err, "unable to get current  component")
	}

	currentComponent := cfg.GetActiveComponent(currentApplication, currentProject)

	return currentComponent, nil

}

func Push(name string, dir string) (string, error) {
	output, err := occlient.StartBuild(name, dir)
	if err != nil {
		return "", errors.Wrap(err, "unable to start build")
	}
	return output, nil
}

// GetComponentType returns type of component in given application and project
func GetComponentType(componentName string, applicationName string, projectName string) (string, error) {
	// filter according to component and application name
	selector := fmt.Sprintf("%s=%s,%s=%s", componentLabel, componentName, application.ApplicationLabel, applicationName)
	ctypes, err := occlient.GetLabelValues(projectName, componentTypeLabel, selector)
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
func List() ([]ComponentInfo, error) {
	// TODO: use project abstaction
	currentProject, err := occlient.GetCurrentProjectName()
	if err != nil {
		return nil, errors.Wrap(err, "unable to list components")
	}

	currentApplication, err := application.GetCurrent()
	if err != nil {
		return nil, errors.Wrap(err, "unable to list components")
	}

	applicationSelector := fmt.Sprintf("%s=%s", application.ApplicationLabel, currentApplication)
	componentNames, err := occlient.GetLabelValues(currentProject, componentLabel, applicationSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to list components")
	}

	components := []ComponentInfo{}

	for _, name := range componentNames {
		ctype, err := GetComponentType(name, currentApplication, currentProject)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to list components")
		}
		components = append(components, ComponentInfo{Name: name, Type: ctype})
	}

	return components, nil
}
