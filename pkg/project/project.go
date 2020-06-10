package project

import (
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/pkg/errors"

	"github.com/openshift/odo/pkg/occlient"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const apiVersion = "odo.dev/v1alpha1"

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
	if projectName == "" {
		return errors.Errorf("no project name given")
	}
	err := client.CreateNewProject(projectName, wait)
	if err != nil {
		return errors.Wrap(err, "unable to create new project")
	}
	if wait {
		err = client.WaitForServiceAccountInNamespace(projectName, "default")
		if err != nil {
			return errors.Wrap(err, "unable to wait for service account")
		}
	}
	return nil
}

// Delete deletes the project with name projectName and returns errors if any
func Delete(client *occlient.Client, projectName string, wait bool) error {
	if projectName == "" {
		return errors.Errorf("no project name given")
	}

	// Delete the requested project
	err := client.DeleteProject(projectName, wait)
	if err != nil {
		return errors.Wrap(err, "unable to delete project")
	}
	return nil
}

// List lists all the projects on the cluster
// returns a list of the projects or the error if any
func List(client *occlient.Client) (ProjectList, error) {
	currentProject := client.GetCurrentProjectName()
	allProjects, err := client.GetProjectNames()
	if err != nil {
		return ProjectList{}, errors.Wrap(err, "cannot get all the projects")
	}
	// Get apps from project
	var projects []Project
	for _, project := range allProjects {
		isActive := false
		if project == currentProject {
			isActive = true
		}
		projects = append(projects, GetMachineReadableFormat(project, isActive))
	}

	return getMachineReadableFormatForList(projects), nil
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
func GetMachineReadableFormat(projectName string, isActive bool) Project {
	return Project{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Project",
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      projectName,
			Namespace: projectName,
		},
		Spec: ProjectSpec{},
		Status: ProjectStatus{
			Active: isActive,
		},
	}
}

// MachineReadableSuccessOutput outputs a success output that includes
// project information and namespace
func MachineReadableSuccessOutput(projectName string, message string) {
	machineOutput := machineoutput.GenericSuccess{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Project",
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      projectName,
			Namespace: projectName,
		},
		Message: message,
	}

	machineoutput.OutputSuccess(machineOutput)
}

// getMachineReadableFormatForList returns application list in machine readable format
func getMachineReadableFormatForList(projects []Project) ProjectList {
	return ProjectList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: apiVersion,
		},
		ListMeta: metav1.ListMeta{},
		Items:    projects,
	}
}
