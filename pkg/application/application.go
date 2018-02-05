package application

import (
	"github.com/pkg/errors"
	"github.com/redhat-developer/ocdev/pkg/occlient"
	log "github.com/sirupsen/logrus"
)

const defaultApplication = "app"

// Create creates a new application and switches to it.

// If no name is provided, the application is named as in the constant
// "defaultApplication".

// If application already exists, this errors out.

// If no project is set, this errors out.
func Create(name string) error {
	// Set default application name if not set
	if len(name) == 0 {
		name = defaultApplication
	}
	if err := occlient.CreateNewProject(name); err != nil {
		return errors.Wrapf(err, "unable to create application: %v", name)
	}
	log.Infof("Switching to application: %v", name)
	return nil
}
