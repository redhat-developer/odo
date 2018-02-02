package application

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/redhat-developer/ocdev/pkg/config"
	"github.com/redhat-developer/ocdev/pkg/occlient"
)

const defaultApplication = "app"

// Create creates a new application and binds it to the current project.

// If no name is provided, the application is named as in the constant
// "defaultApplication".

// If application already exists, this errors out.

// If no project is set, this errors out.
func Create(name string) error {
	// Get current project name
	project, err := occlient.GetCurrentProjectName()
	if err != nil {
		return errors.Wrap(err, "unable to get current project's name")
	}

	// Set default application name if not set
	if len(name) == 0 {
		name = defaultApplication
	}

	app := config.Application{
		Name:    name,
		Project: project,
	}

	ocdevConfig, err := config.New()
	if err != nil {
		return errors.Wrap(err, "error getting config")
	}

	// Check if application exists
	if ocdevConfig.ApplicationExists(&app) {
		return fmt.Errorf("application %v already exists in project %v", app.Name, app.Project)
	}

	// Add application to config
	err = ocdevConfig.AddApplication(&app)
	if err != nil {
		return errors.Wrap(err, "unable to add application to config")
	}

	return nil
}
