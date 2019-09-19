package project

import (
	"github.com/openshift/odo/pkg/application"
	"github.com/pkg/errors"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

func DescribeProjects(client *occlient.Client) (ProjectList, error) {
	allProjects, err := client.GetProjectNames()
	if err != nil {
		return ProjectList{}, errors.Wrap(err, "cannot get all the projects")
	}

	// Go through each project
	var projects []Project
	for _, project := range allProjects {
		projects = append(projects, GetMachineReadableFormat(client, project))
	}

	return ProjectList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ProjectList",
			APIVersion: "odo.openshift.io/v1alpha1",
		},
		ListMeta: metav1.ListMeta{},
		Items:    projects,
	}, nil
}

// Exists Checks whether a project with the given name exists or not
// projectName is the project name to perform check for
// The first returned parameter is a bool indicating if a project with the given name already exists or not
// The second returned parameter is the error that might occurs while execution
func Exists(client *occlient.Client, projectName string) (bool, error) {
	project, err := client.GetProject(projectName)
	if err != nil || project == nil {
		return false, err
	}

	return true, nil
}

// GetMachineReadableFormat gathers the readable format and output a Project struct
// for json to marshal
func GetMachineReadableFormat(client *occlient.Client, projectName string) Project {
	isActive := false

	// Retrieve the current project
	currentProject := client.GetCurrentProjectName()
	if projectName == currentProject {
		isActive = true
	}

	// Get a list of apps
	apps, _ := application.ListInProject(client)

	return Project{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Project",
			APIVersion: "odo.openshift.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      projectName,
			Namespace: projectName,
		},
		Spec: ProjectSpec{
			Applications: apps,
		},
		Status: ProjectStatus{
			Active: isActive,
		},
	}
}
