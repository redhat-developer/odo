package application

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/redhat-developer/odo/pkg/util"
)

const (
	appPrefixMaxLen   = 12
	appNameMaxRetries = 3
	appAPIVersion     = "odo.openshift.io/v1alpha1"
	appKind           = "app"
	appList           = "List"
)

// GetDefaultAppName returns randomly generated application name with unique configurable prefix suffixed by a randomly generated string which can be used as a default name in case the user doesn't provide a name.
func GetDefaultAppName(existingApps []preference.ApplicationInfo) (string, error) {
	var appName string
	var existingAppNames []string

	// Get list of app names
	for _, app := range existingApps {
		existingAppNames = append(existingAppNames, app.Name)
	}

	// Get the desired app name prefix from odo config
	cfg, err := preference.New()
	if err != nil {
		return "", errors.Wrap(err, "unable to fetch config")
	}

	// If there's no prefix in config file or it is equal to $DIR, use safe default which is the name of current directory
	if cfg.OdoSettings.NamePrefix == nil || *cfg.OdoSettings.NamePrefix == "" {
		prefix, err := component.GetComponentDir("", occlient.NONE)
		if err != nil {
			return "", errors.Wrap(err, "unable to generate random app name")
		}
		appName, err = util.GetRandomName(prefix, appPrefixMaxLen, existingAppNames, appNameMaxRetries)
		if err != nil {
			return "", errors.Wrap(err, "unable to generate random app name")
		}
	} else {
		appName, err = util.GetRandomName(*cfg.OdoSettings.NamePrefix, appPrefixMaxLen, existingAppNames, appNameMaxRetries)
	}
	if err != nil {
		return "", errors.Wrap(err, "unable to generate random app name")
	}
	return util.GetDNS1123Name(appName), nil
}

// Create a new application
func Create(client *occlient.Client, appName string) error {

	exists, _ := Exists(client, appName)

	if exists {
		return fmt.Errorf("unable to create new application, %s application already exists", appName)
	}

	cfg, err := preference.New()
	if err != nil {
		return errors.Wrap(err, "unable to create new application")
	}

	err = cfg.AddApplication(appName, client.Namespace)
	if err != nil {
		return errors.Wrap(err, "unable to create new application")
	}
	return nil
}

// List all applications in current project
func List(client *occlient.Client) ([]preference.ApplicationInfo, error) {
	return ListInProject(client, client.Namespace)
}

// ListInProject lists all applications in given project
// Queries cluster and config file.
// Shows also empty applications (empty applications are those that are just
// mentioned in config file but don't have any object associated with it on cluster).
// ListInProject lists all applications in given project
// Queries cluster and config file.
// Shows also empty applications (empty applications are those that are just
// mentioned in config file but don't have any object associated with it on cluster).
func ListInProject(client *occlient.Client, projectName string) ([]preference.ApplicationInfo, error) {
	var applications []preference.ApplicationInfo

	cfg, err := preference.New()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create new application")
	}

	// All applications of the current project from config file
	for i := range cfg.ActiveApplications {
		if cfg.ActiveApplications[i].Project == projectName {
			applications = append(applications, cfg.ActiveApplications[i])
		}
	}

	// Get applications from cluster
	appNames, err := client.GetLabelValues(applabels.ApplicationLabel, applabels.ApplicationLabel)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list applications")
	}

	for _, name := range appNames {
		// skip applications that are already in the list (they were mentioned in config file)
		found := false
		for _, app := range applications {
			if app.Project == projectName && app.Name == name {
				found = true
			}
		}
		if !found {
			applications = append(applications, preference.ApplicationInfo{
				Name: name,
				// if this application is not in config file, it can't be active
				Active:  false,
				Project: projectName,
			})
		}
	}

	return applications, nil
}

// Delete deletes the given application
func Delete(client *occlient.Client, name string) error {
	glog.V(4).Infof("Deleting application %s", name)

	labels := applabels.GetLabels(name, false)

	// delete application from cluster
	err := client.Delete(labels)
	if err != nil {
		return errors.Wrapf(err, "unable to delete application %s", name)
	}

	// delete from config
	cfg, err := preference.New()
	if err != nil {
		return errors.Wrapf(err, "unable to delete application %s", name)
	}

	err = cfg.DeleteApplication(name, client.Namespace)
	if err != nil {
		return errors.Wrapf(err, "unable to delete application %s", name)
	}

	return nil
}

// GetCurrent returns currently active application.
// If no application is active this functions returns empty string
func GetCurrent(projectName string) (string, error) {
	cfg, err := preference.New()
	if err != nil {
		return "", errors.Wrap(err, "unable to get active application")
	}

	app := cfg.GetActiveApplication(projectName)
	return app, nil
}

// GetCurrentOrGetCreateSetDefault returns currently active application.
// If no application is active, a defaultApplication is created and set as
// default as well.
// Use this carefully only in places where user expects the state to be altered
// Do not use for read operations like get, list; only for write operations like
// create
func GetCurrentOrGetCreateSetDefault(client *occlient.Client) (string, error) {
	currentApp, err := GetCurrent(client.Namespace)
	if err != nil {
		return "", errors.Wrap(err, "unable to get active application")
	}
	// if no Application is active use default
	if currentApp == "" {
		// get default application name
		defaultName, err := GetDefaultAppName([]preference.ApplicationInfo{})
		if err != nil {
			return "", errors.Wrap(err, "unable to fetch/create an application to set as active")
		}
		currentApp = defaultName
		// create if default application does not exist
		exists, _ := Exists(client, currentApp)
		if !exists {
			if err := Create(client, currentApp); err != nil {
				return "", errors.Wrapf(err, "unable to create app %v", currentApp)
			}
		}
		// set default application as the current application
		if err := SetCurrent(client, currentApp); err != nil {
			return "", errors.Wrapf(err, "unable to set %v as the current application", currentApp)
		}
	}
	return currentApp, nil
}

// SetCurrent set application as active
func SetCurrent(client *occlient.Client, appName string) error {
	// also sets project to the given project
	err := project.SetCurrent(client, client.Namespace)
	if err != nil {
		return errors.Wrap(err, "unable to set current project")
	}
	glog.V(4).Infof("Setting application %s as current.\n", appName)

	cfg, err := preference.New()
	if err != nil {
		return errors.Wrap(err, "unable to set current application")
	}

	exists, err := Exists(client, appName)
	if err != nil {
		return errors.Wrap(err, "unable to set current application")
	}
	if !exists {
		return fmt.Errorf("application %s doesn't exist", appName)
	}

	// There might be a situation where application is not defined in local config
	// but it is present in OpenShift cluster. This situation can happen for example if user deleted config file.
	// In that case we need to add application back to the the config before we set it as active.
	found := false
	for _, cfgApp := range cfg.ActiveApplications {
		if cfgApp.Project == client.Namespace && cfgApp.Name == appName {
			found = true
			break
		}
	}
	if !found {
		err := cfg.AddApplication(appName, client.Namespace)
		if err != nil {
			return errors.Wrap(err, "unable to add application")
		}
	}

	err = cfg.SetActiveApplication(appName, client.Namespace)
	if err != nil {
		return errors.Wrap(err, "unable to set current application")
	}

	return nil
}

// Exists returns true if given application (appName) exists
func Exists(client *occlient.Client, appName string) (bool, error) {
	apps, err := List(client)
	if err != nil {
		return false, errors.Wrap(err, "unable to list applications")
	}
	for _, app := range apps {
		if app.Name == appName {
			return true, nil
		}
	}
	return false, errors.Errorf("application %v does not exist in project %v", appName, client.Namespace)
}

// GetMachineReadableFormat returns resource information in machine readable format
func GetMachineReadableFormat(client *occlient.Client, appName string, projectName string, active bool) App {
	componentList, _ := component.List(client, appName)
	var compList []string
	for _, comp := range componentList.Items {
		compList = append(compList, comp.Name)
	}
	appDef := App{
		TypeMeta: metav1.TypeMeta{
			Kind:       appKind,
			APIVersion: appAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: projectName,
		},
		Spec: AppSpec{
			Components: compList,
		},
		Status: AppStatus{
			Active: active,
		},
	}
	return appDef
}

// GetMachineReadableFormatForList returns application list in machine readable format
func GetMachineReadableFormatForList(apps []App) AppList {
	return AppList{
		TypeMeta: metav1.TypeMeta{
			Kind:       appList,
			APIVersion: appAPIVersion,
		},
		ListMeta: metav1.ListMeta{},
		Items:    apps,
	}
}
