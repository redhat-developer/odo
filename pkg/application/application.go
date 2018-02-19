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

// ApplicationInfo holds all important information about applications
type ApplicationInfo struct {
	// name of the application
	Name string
	// is this application active? Only one application can be active at the time
	Active bool
	// name of the openshift project this application belongs to
	Project string
}

// Create a new application and set is as active.
// If application already exists, this errors out.
// If no project is set, this errors out.
func Create(name string) error {
	// TODO: right now this assumes that there is a current project in openshift
	// when we have project support in ocdev, this should call project.GetCurrent()
	// TODO: use project abstraction
	project, err := occlient.GetCurrentProjectName()
	if err != nil {
		return errors.Wrap(err, "unable to create new application")
	}

	cfg, err := config.New()
	if err != nil {
		return errors.Wrap(err, "unable to create new application")
	}

	err = cfg.SetActiveApplication(name, project)
	if err != nil {
		return errors.Wrap(err, "unable to create new application")
	}

	return nil
}

// List all application in current project
// shows also empty applications (empty applications are those that are just
//  mentioned in config but don't have any object associated with it on cluster)
func List() ([]ApplicationInfo, error) {
	// TODO: use project abstaction
	project, err := occlient.GetCurrentProjectName()
	if err != nil {
		return nil, errors.Wrap(err, "unable to list applications")
	}

	appNames, err := occlient.GetLabelValues(project, ApplicationLabel)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list applications")
	}

	activeApplication, err := GetCurrent()
	if err != nil {
		return nil, errors.Wrap(err, "unable to list applications")
	}

	applications := []ApplicationInfo{}

	for _, name := range appNames {
		active := false
		if activeApplication == name {
			active = true
		}
		applications = append(applications, ApplicationInfo{
			Name:    name,
			Active:  active,
			Project: project,
		})
	}

	cfg, err := config.New()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create new application")
	}

	// Application can be created but empty.
	// In that case it won't show in the list above, as there are no objects with application
	// label in the cluster. Only place where this application is mentioned is local config.
	activeApp := cfg.GetActiveApplication(project)
	if activeApp != "" {
		applications = append(applications, ApplicationInfo{
			Name:    activeApp,
			Active:  true,
			Project: project,
		})
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
