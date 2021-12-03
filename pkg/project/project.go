package project

import (
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/kclient"
)

// SetCurrent sets projectName as the current project
func SetCurrent(client kclient.ClientInterface, projectName string) error {
	err := client.SetCurrentNamespace(projectName)
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
func Create(client kclient.ClientInterface, projectName string, wait bool) error {
	if projectName == "" {
		return errors.Errorf("no project name given")
	}

	projectSupport, err := client.IsProjectSupported()
	if err != nil {
		return errors.Wrap(err, "unable to detect project support")
	}
	if projectSupport {
		err = client.CreateNewProject(projectName, wait)
		if err != nil {
			return errors.Wrap(err, "unable to create new project")
		}

	} else {
		_, err = client.CreateNamespace(projectName)
		if err != nil {
			return errors.Wrap(err, "unable to create new project")
		}
	}

	if wait {
		err = client.WaitForServiceAccountInNamespace(projectName, "default")
		if err != nil {
			return errors.Wrap(err, "unable to wait for service account")
		}
	}
	return nil
}

// Delete deletes the project (the `project` resource if supported, or directly the `namespace`)
// with the name projectName and returns an error if any
func Delete(client kclient.ClientInterface, projectName string, wait bool) error {
	if projectName == "" {
		return errors.Errorf("no project name given")
	}

	projectSupport, err := client.IsProjectSupported()
	if err != nil {
		return errors.Wrap(err, "unable to detect project support")
	}

	if projectSupport {
		// Delete the requested project
		err := client.DeleteProject(projectName, wait)
		if err != nil {
			return errors.Wrapf(err, "unable to delete project %s", projectName)
		}
	} else {
		err := client.DeleteNamespace(projectName, wait)
		if err != nil {
			return errors.Wrapf(err, "unable to delete namespace %s", projectName)
		}
	}
	return nil
}

// List all the projects on the cluster and returns an error if any
func List(client kclient.ClientInterface) (ProjectList, error) {
	currentProject := client.GetCurrentNamespace()

	projectSupport, err := client.IsProjectSupported()
	if err != nil {
		return ProjectList{}, errors.Wrap(err, "unable to detect project support")
	}

	var allProjects []string
	if projectSupport {
		allProjects, err = client.ListProjectNames()
		if err != nil {
			return ProjectList{}, errors.Wrap(err, "cannot get all the projects")
		}
	} else {
		allProjects, err = client.GetNamespaces()
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
func Exists(client kclient.ClientInterface, projectName string) (bool, error) {
	projectSupport, err := client.IsProjectSupported()
	if err != nil {
		return false, errors.Wrap(err, "unable to detect project support")
	}

	if projectSupport {
		project, err := client.GetProject(projectName)
		if err != nil || project == nil {
			return false, err
		}
	} else {
		namespace, err := client.GetNamespace(projectName)
		if err != nil || namespace == nil {
			return false, err
		}
	}

	return true, nil
}
