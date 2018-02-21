package application

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/redhat-developer/ocdev/pkg/config"
	"github.com/redhat-developer/ocdev/pkg/occlient"
)

const defaultApplication = "app"

// ApplicationLabel is label key that is used to group all object that belong to one application
// It should be save to use just this label to filter application
const ApplicationLabel = "app.kubernetes.io/name"

// AdditionalApplicationLabels additional labels that are applied to all objects belonging to one application
// Those labels are not used for filtering or grouping, they are used just when creating and they are mend to be used by other tools
var AdditionalApplicationLabels = []string{
	// OpenShift Web console uses this label for grouping
	"app",
}

// Create a new application and set is as active.
// If application already exists, this errors out.
// If no project is set, this errors out.
func Create(name string) error {
	err := SetCurrent(name)
	if err != nil {
		return errors.Wrapf(err, "unable to create new application")
	}

	return nil
}

// List all application in current project
// shows also empty applications (empty applications are those that are just
// mentioned in config but don't have any object associated with it on cluster)
func List() ([]config.ApplicationInfo, error) {
	applications := []config.ApplicationInfo{}

	// TODO: use project abstaction
	project, err := occlient.GetCurrentProjectName()
	if err != nil {
		return nil, errors.Wrap(err, "unable to list applications")
	}

	cfg, err := config.New()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create new application")
	}

	// All applications from config file
	applications = append(applications, cfg.ActiveApplications...)

	// Get applications from cluster
	appNames, err := occlient.GetLabelValues(project, ApplicationLabel)
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
func Delete(name string) error {
	// if err := occlient.DeleteProject(name); err != nil {
	// 	return errors.Wrapf(err, "unable to delete application: %v", name)
	// }
	// TODO: implement
	return fmt.Errorf("NOT IMPLEMENTED")
}

// GetCurrent application if no application is active it returns defaultApplication name
func GetCurrent() (string, error) {
	// TODO: use project abstaction
	project, err := occlient.GetCurrentProjectName()
	if err != nil {
		return "", errors.Wrap(err, "unable to get active application")
	}

	cfg, err := config.New()
	if err != nil {
		return "", errors.Wrap(err, "unable to get active application")
	}

	app := cfg.GetActiveApplication(project)
	// if no Application is active use default
	if app == "" {
		app = defaultApplication
	}

	return app, nil
}

// SetCurrent set application as active
func SetCurrent(name string) error {
	// TODO: right now this assumes that there is a current project in openshift
	// when we have project support in ocdev, this should call project.GetCurrent()
	// TODO: use project abstraction
	project, err := occlient.GetCurrentProjectName()
	if err != nil {
		return errors.Wrap(err, "unable to get active application")
	}

	cfg, err := config.New()
	if err != nil {
		return errors.Wrap(err, "unable to get active application")
	}

	err = cfg.SetActiveApplication(name, project)
	if err != nil {
		return errors.Wrap(err, "unable to create new application")
	}

	return nil
}
