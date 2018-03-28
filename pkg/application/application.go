package application

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/redhat-developer/ocdev/pkg/config"
	"github.com/redhat-developer/ocdev/pkg/occlient"
	log "github.com/sirupsen/logrus"
)

// ApplicationLabel is label key that is used to group all object that belong to one application
// It should be save to use just this label to filter application
const ApplicationLabel = "app.kubernetes.io/name"

// AdditionalApplicationLabels additional labels that are applied to all objects belonging to one application
// Those labels are not used for filtering or grouping, they are used just when creating and they are mend to be used by other tools
var AdditionalApplicationLabels = []string{
	// OpenShift Web console uses this label for grouping
	"app",
}

// getDefaultAppName returns application name to be used as a default name in the case where users doesn't provide a name
// In future this function should generate name with uniq suffix (app-xy1h), because there might be multiple applications.
func getDefaultAppName() string {
	return "app"
}

// GetLabels return labels that identifies given application
// additional labels are used only when creating object
// if you are creating something use additional=true
// if you need labels to filter component than use additional=false
func GetLabels(application string, additional bool) (map[string]string, error) {
	labels := map[string]string{
		ApplicationLabel: application,
	}

	if additional {
		for _, additionalLabel := range AdditionalApplicationLabels {
			labels[additionalLabel] = application
		}
	}

	return labels, nil
}

// Create a new application
func Create(client *occlient.Client, applicationName string) error {
	// TODO: use project abstraction
	project := client.GetCurrentProjectName()

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

// List all application in current project
// Queries cluster and configfile.
// Shows also empty applications (empty applications are those that are just
// mentioned in config but don't have any object associated with it on cluster).
func List(client *occlient.Client) ([]config.ApplicationInfo, error) {
	applications := []config.ApplicationInfo{}

	// TODO: use project abstaction
	project := client.GetCurrentProjectName()

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
	appNames, err := client.GetLabelValues(project, ApplicationLabel, ApplicationLabel)
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
	log.Debug("Deleting application %s", name)

	labels, err := GetLabels(name, false)
	if err != nil {
		return errors.Wrapf(err, "unable to delete application %s", name)
	}

	// delete application from cluster
	output, err := client.Delete("all", "", labels)
	if err != nil {
		return errors.Wrapf(err, "unable to delete application %s", name)
	}
	log.Debug("deleted from cluster: \n", output)

	// delete from config
	cfg, err := config.New()
	if err != nil {
		return errors.Wrapf(err, "unable to delete application %s", name)
	}
	project := client.GetCurrentProjectName()

	err = cfg.DeleteApplication(name, project)
	if err != nil {
		return errors.Wrapf(err, "unable to delete application %s", name)
	}
	log.Debug("deleted from config: \n", output)

	return nil
}

// GetCurrent returns currently active application.
// If no application is active this functions returns empty string
func GetCurrent(client *occlient.Client) (string, error) {
	// TODO: use project abstaction
	project := client.GetCurrentProjectName()

	cfg, err := config.New()
	if err != nil {
		return "", errors.Wrap(err, "unable to get active application")
	}

	app := cfg.GetActiveApplication(project)
	return app, nil
}

// GetCurrentOrDefault returns currently active application.
// If no application is active returns defaultApplication name
func GetCurrentOrDefault(client *occlient.Client) (string, error) {
	currentApp, err := GetCurrent(client)
	if err != nil {
		return "", errors.Wrap(err, "unable to get active application")
	}
	// if no Application is active use default
	if currentApp == "" {
		currentApp = getDefaultAppName()
	}
	return currentApp, nil
}

// SetCurrent set application as active
func SetCurrent(client *occlient.Client, name string) error {
	// TODO: right now this assumes that there is a current project in openshift
	// when we have project support in ocdev, this should call project.GetCurrent()
	// TODO: use project abstraction
	log.Debugf("Setting application %s as current.\n", name)

	project := client.GetCurrentProjectName()

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
	// but it is present in OpenShift cluster. This situation can happen for example if user delted config file.
	// In this case we need to add application back to the the config  before we set is as active.
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
