package project

import (
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const apiVersion = "odo.dev/v1alpha1"

// GetCurrent returns the name of the current project
func GetCurrent(context *genericclioptions.Context) string {
	return context.KClient.GetCurrentNamespace()
}

// SetCurrent sets projectName as the current project
func SetCurrent(context *genericclioptions.Context, projectName string) error {
	err := context.KClient.SetCurrentNamespace(projectName)
	if err != nil {
		return errors.Wrap(err, "unable to set current project to"+projectName)
	}
	return nil
}

// Create a new project, either by creating a `project.openshift.io` resource if supported by the cluster
// (which will trigger the creation of a namespace),
// or by creating directly a `namespace` resource.
// With the `wait` flag, the function will wait for the `default` service account
// to be created in the namespace before to return.
func Create(context *genericclioptions.Context, projectName string, wait bool) error {
	if projectName == "" {
		return errors.Errorf("no project name given")
	}

	projectSupport, err := context.Client.IsProjectSupported()
	if err != nil {
		return errors.Wrap(err, "unable to detect project support")
	}
	if projectSupport {
		err := context.Client.CreateNewProject(projectName, wait)
		if err != nil {
			return errors.Wrap(err, "unable to create new project")
		}

	} else {
		_, err := context.KClient.CreateNamespace(projectName)
		if err != nil {
			return errors.Wrap(err, "unable to create new project")
		}
	}

	if wait {
		err = context.KClient.WaitForServiceAccountInNamespace(projectName, "default")
		if err != nil {
			return errors.Wrap(err, "unable to wait for service account")
		}
	}
	return nil
}

// Delete deletes the project (the `project` resource if supported, or directly the `namespace`)
// with the name projectName and returns an error if any
func Delete(context *genericclioptions.Context, projectName string, wait bool) error {
	if projectName == "" {
		return errors.Errorf("no project name given")
	}

	projectSupport, err := context.Client.IsProjectSupported()
	if err != nil {
		return errors.Wrap(err, "unable to detect project support")
	}

	if projectSupport {
		// Delete the requested project
		err := context.Client.DeleteProject(projectName, wait)
		if err != nil {
			return errors.Wrapf(err, "unable to delete project %s", projectName)
		}
	} else {
		err := context.KClient.DeleteNamespace(projectName, wait)
		if err != nil {
			return errors.Wrapf(err, "unable to delete namespace %s", projectName)
		}
	}
	return nil
}

// List all the projects on the cluster on a machine readable format and returns an error if any
func List(context *genericclioptions.Context) (ProjectList, error) {
	currentProject := context.KClient.GetCurrentNamespace()

	projectSupport, err := context.Client.IsProjectSupported()
	if err != nil {
		return ProjectList{}, errors.Wrap(err, "unable to detect project support")
	}

	var allProjects []string
	if projectSupport {
		allProjects, err = context.Client.ListProjectNames()
		if err != nil {
			return ProjectList{}, errors.Wrap(err, "cannot get all the projects")
		}
	} else {
		allProjects, err = context.KClient.GetNamespaces()
		if err != nil {
			return ProjectList{}, errors.Wrap(err, "cannot get all the namespaces")
		}
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

// Exists checks whether a project with the name `projectName` exists and returns an error if any
func Exists(context *genericclioptions.Context, projectName string) (bool, error) {
	projectSupport, err := context.Client.IsProjectSupported()
	if err != nil {
		return false, errors.Wrap(err, "unable to detect project support")
	}

	if projectSupport {
		project, err := context.Client.GetProject(projectName)
		if err != nil || project == nil {
			return false, err
		}
	} else {
		namespace, err := context.KClient.GetNamespace(projectName)
		if err != nil || namespace == nil {
			return false, err
		}
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
