package project

import (
	"github.com/openshift/odo/v2/pkg/odo/genericclioptions"
	"github.com/pkg/errors"
)

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

// List all the projects on the cluster and returns an error if any
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

	projects := make([]Project, len(allProjects))
	for i, project := range allProjects {
		isActive := project == currentProject
		projects[i] = NewProject(project, isActive)
	}

	return NewProjectList(projects), nil
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
