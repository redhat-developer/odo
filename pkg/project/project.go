package project

import (
	"github.com/pkg/errors"
	"github.com/redhat-developer/ocdev/pkg/occlient"
)

func GetCurrent(client *occlient.Client) (string, error) {
	// TODO: use project abstaction
	project, err := client.GetCurrentProjectName()
	if err != nil {
		return "", errors.Wrap(err, "unable to get current project")
	}

	return project, nil
}

func SetCurrent(client *occlient.Client, project string) error {
	err := client.SetCurrentProject(project)
	if err != nil {
		return errors.Wrap(err, "unable to set current project to"+project)
	}
	return nil
}

func CreateProject(client *occlient.Client, projectName string) error {
	err := client.CreateNewProject(projectName)
	if err != nil {
		return errors.Wrap(err, "unable to create new project")
	}
	return nil
}
