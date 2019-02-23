package project

import (
	"github.com/pkg/errors"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
)

// ApplicationInfo holds information about one project
type ProjectInfo struct {
	// Name of the project
	Name string
	// is this project active?
	Active bool
}

// GetCurrent return current project
func GetCurrent(client *occlient.Client) string {
	project := client.GetCurrentProjectName()
	return project
}

// SetCurrent sets the projectName as current project
func SetCurrent(client *occlient.Client, projectName string) error {
	err := client.SetCurrentProject(projectName)
	if err != nil {
		return errors.Wrap(err, "unable to set current project to"+projectName)
	}
	return nil
}

func Create(client *occlient.Client, projectName string, wait bool) error {
	err := client.CreateNewProject(projectName, wait)
	if err != nil {
		return errors.Wrap(err, "unable to create new project")
	}
	return nil
}

// Delete deletes the project with name projectName and returns errors if any
func Delete(client *occlient.Client, projectName string) error {
	// Loading spinner
	s := log.Spinnerf("Deleting project %s", projectName)
	defer s.End(false)

	// Delete the requested project
	err := client.DeleteProject(projectName)
	if err != nil {
		return errors.Wrap(err, "unable to delete project")
	}

	s.End(true)
	return nil
}

func List(client *occlient.Client) ([]ProjectInfo, error) {
	currentProject := client.GetCurrentProjectName()
	allProjects, err := client.GetProjectNames()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get all the projects")
	}
	var projects []ProjectInfo
	for _, project := range allProjects {
		isActive := false
		if project == currentProject {
			isActive = true
		}
		projects = append(projects, ProjectInfo{
			Name:   project,
			Active: isActive,
		})
	}
	return projects, nil
}

// Checks whether a project with the given name exists or not
// projectName is the project name to perform check for
// The first returned parameter is a bool indicating if a project with the given name already exists or not
// The second returned parameter is the error that might occurs while execution
func Exists(client *occlient.Client, projectName string) (bool, error) {
	projects, err := List(client)
	if err != nil {
		return false, errors.Wrap(err, "unable to get the project list")
	}
	for _, project := range projects {
		if project.Name == projectName {
			return true, nil
		}
	}
	return false, nil
}
