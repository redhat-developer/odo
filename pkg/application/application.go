package application

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	"github.com/redhat-developer/odo/pkg/config"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/redhat-developer/odo/pkg/util"
)

const (
	appPrefixMaxLen   = 12
	appNameMaxRetries = 3
)

// GetDefaultAppName returns randomly generated application name with unique configurable prefix suffixed by a randomly generated string which canbe used as a default name in case the user doesn't provide a name.
func GetDefaultAppName(existingApps []config.ApplicationInfo) (string, error) {
	var appName string
	var existingAppNames []string

	// Get list of app names
	for _, app := range existingApps {
		existingAppNames = append(existingAppNames, app.Name)
	}

	// Get the desired app name prefix from odo config
	cfg, err := config.New()
	if err != nil {
		return "", errors.Wrap(err, "unable to fetch config")
	}

	// If there's no prefix in config file or it is equal to $DIR, use safe default which is the name of current directory
	if cfg.OdoSettings.Prefix == nil || *cfg.OdoSettings.Prefix == config.ConfigPrefixDir {
		prefix, err := util.GetComponentDir("", util.NONE)
		if err != nil {
			return "", errors.Wrap(err, "unable to generate random app name")
		}
		appName, err = util.GetRandomName(prefix, appPrefixMaxLen, existingAppNames, appNameMaxRetries)
		if err != nil {
			return "", errors.Wrap(err, "unable to generate random app name")
		}
	} else {
		appName, err = util.GetRandomName(*cfg.OdoSettings.Prefix, appPrefixMaxLen, existingAppNames, appNameMaxRetries)
	}
	if err != nil {
		return "", errors.Wrap(err, "unable to generate random app name")
	}
	return util.GetDNS1123Name(appName), nil
}

// Create a new application
func Create(client *occlient.Client, applicationName string) error {
	project := project.GetCurrent(client)

	exists, err := Exists(client, applicationName)
	if err != nil {
		return errors.Wrap(err, "unable to create new application")
	}
	if exists {
		return fmt.Errorf("unable to create new application, %s application already exists", applicationName)
	}

	cfg, err := config.New()
	if err != nil {
		return errors.Wrap(err, "unable to create new application")
	}

	err = cfg.AddApplication(applicationName, project)
	if err != nil {
		return errors.Wrap(err, "unable to create new application")
	}
	return nil
}

// List all applications in current project
func List(client *occlient.Client) ([]config.ApplicationInfo, error) {
	return ListInProject(client, project.GetCurrent(client))
}

// ListInProject lists all applications in given project
// Queries cluster and config file.
// Shows also empty applications (empty applications are those that are just
// mentioned in config file but don't have any object associated with it on cluster).
func ListInProject(client *occlient.Client, project string) ([]config.ApplicationInfo, error) {
	applications := []config.ApplicationInfo{}

	cfg, err := config.New()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create new application")
	}

	// All applications of the current project from config file
	for i := range cfg.ActiveApplications {
		if cfg.ActiveApplications[i].Project == project {
			applications = append(applications, cfg.ActiveApplications[i])
		}
	}

	// Get applications from cluster
	appNames, err := client.GetLabelValues(project, applabels.ApplicationLabel, applabels.ApplicationLabel)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list applications")
	}

	for _, name := range appNames {
		// skip applications that are already in the list (they were mentioned in config file)
		found := false
		for _, app := range applications {
			if app.Project == project && app.Name == name {
				found = true
			}
		}
		if !found {
			applications = append(applications, config.ApplicationInfo{
				Name: name,
				// if this application is not in config file, it can't be active
				Active:  false,
				Project: project,
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
	cfg, err := config.New()
	if err != nil {
		return errors.Wrapf(err, "unable to delete application %s", name)
	}

	project := project.GetCurrent(client)

	err = cfg.DeleteApplication(name, project)
	if err != nil {
		return errors.Wrapf(err, "unable to delete application %s", name)
	}

	return nil
}

// GetCurrent returns currently active application.
// If no application is active this functions returns empty string
func GetCurrent(client *occlient.Client) (string, error) {
	project := project.GetCurrent(client)

	cfg, err := config.New()
	if err != nil {
		return "", errors.Wrap(err, "unable to get active application")
	}

	app := cfg.GetActiveApplication(project)
	return app, nil
}

// GetCurrentOrGetCreateSetDefault returns currently active application.
// If no application is active, a defaultApplication is created and set as
// default as well.
// Use this carefully only in places where user expects the state to be altered
// Do not use for read operations like get, list; only for write operations like
// create
func GetCurrentOrGetCreateSetDefault(client *occlient.Client) (string, error) {
	currentApp, err := GetCurrent(client)
	if err != nil {
		return "", errors.Wrap(err, "unable to get active application")
	}

	// if no Application is active use default
	if currentApp == "" {
		// get default application name
		defaultName, err := GetDefaultAppName([]config.ApplicationInfo{})
		if err != nil {
			return "", errors.Wrap(err, "unable to fetch/create an application to set as active")
		}
		currentApp = defaultName
		// create if default application does not exist
		exists, err := Exists(client, currentApp)
		if err != nil {
			return "", errors.Wrapf(err, "unable to check if app %v exists", currentApp)
		}
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
func SetCurrent(client *occlient.Client, name string) error {
	glog.V(4).Infof("Setting application %s as current.\n", name)

	project := project.GetCurrent(client)

	cfg, err := config.New()
	if err != nil {
		return errors.Wrap(err, "unable to set current application")
	}

	exists, err := Exists(client, name)
	if err != nil {
		return errors.Wrap(err, "unable to set current application")
	}
	if !exists {
		return fmt.Errorf("application %s doesn't exist", name)
	}

	// There might be a situation where application is not defined in local config
	// but it is present in OpenShift cluster. This situation can happen for example if user deleted config file.
	// In that case we need to add application back to the the config before we set it as active.
	found := false
	for _, cfgApp := range cfg.ActiveApplications {
		if cfgApp.Project == project && cfgApp.Name == name {
			found = true
			break
		}
	}
	if !found {
		cfg.AddApplication(name, project)
	}

	err = cfg.SetActiveApplication(name, project)
	if err != nil {
		return errors.Wrap(err, "unable to set current application")
	}

	return nil
}

// Exists returns true if given application name exist
func Exists(client *occlient.Client, name string) (bool, error) {
	apps, err := List(client)
	if err != nil {
		return false, errors.Wrap(err, "unable to list applications")
	}
	for _, app := range apps {
		if app.Name == name {
			return true, nil
		}
	}
	return false, nil
}
