package project

import (
	"errors"
	"fmt"

	"github.com/redhat-developer/odo/pkg/kclient"
)

type kubernetesClient struct {
	client kclient.ClientInterface
}

var _ Client = (*kubernetesClient)(nil)

func NewClient(client kclient.ClientInterface) Client {
	return kubernetesClient{
		client: client,
	}
}

// SetCurrent sets projectName as the current project
func (o kubernetesClient) SetCurrent(projectName string) error {
	err := o.client.SetCurrentNamespace(projectName)
	if err != nil {
		return fmt.Errorf("unable to set current project to %s: %w", projectName, err)
	}
	return nil
}

// Create a new project, either by creating a `project.openshift.io` resource if supported by the cluster
// (which will trigger the creation of a namespace),
// or by creating directly a `namespace` resource.
// With the `wait` flag, the function will wait for the `default` service account
// to be created in the namespace before returning
func (o kubernetesClient) Create(projectName string, wait bool) error {
	if projectName == "" {
		return errors.New("no project name given")
	}

	projectSupport, err := o.client.IsProjectSupported()
	if err != nil {
		return fmt.Errorf("unable to detect project support: %w", err)
	}

	if projectSupport {
		err = o.client.CreateNewProject(projectName, wait)

	} else {
		_, err = o.client.CreateNamespace(projectName)
	}
	if err != nil {
		return fmt.Errorf("unable to create new project: %w", err)
	}

	if wait {
		err = o.client.WaitForServiceAccountInNamespace(projectName, "default")
		if err != nil {
			return fmt.Errorf("unable to wait for service account: %w", err)
		}
	}
	return nil
}

// Delete deletes the project (the `project` resource if supported, or directly the `namespace`)
// with the name projectName and returns an error if any
func (o kubernetesClient) Delete(projectName string, wait bool) error {
	if projectName == "" {
		return errors.New("no project name given")
	}

	projectSupport, err := o.client.IsProjectSupported()
	if err != nil {
		return fmt.Errorf("unable to detect project support: %w", err)
	}

	if projectSupport {
		err = o.client.DeleteProject(projectName, wait)
	} else {
		err = o.client.DeleteNamespace(projectName, wait)
	}
	if err != nil {
		return fmt.Errorf("unable to delete project %q: %w", projectName, err)
	}
	return nil
}

// List all the projects on the cluster and returns an error if any
func (o kubernetesClient) List() (ProjectList, error) {
	currentProject := o.client.GetCurrentNamespace()

	projectSupport, err := o.client.IsProjectSupported()
	if err != nil {
		return ProjectList{}, fmt.Errorf("unable to detect project support: %w", err)
	}

	var allProjects []string
	if projectSupport {
		allProjects, err = o.client.ListProjectNames()
	} else {
		allProjects, err = o.client.GetNamespaces()
	}
	if err != nil {
		return ProjectList{}, fmt.Errorf("cannot get all the projects: %w", err)
	}

	projects := make([]Project, len(allProjects))
	for i, project := range allProjects {
		isActive := project == currentProject
		projects[i] = NewProject(project, isActive)
	}

	return NewProjectList(projects), nil
}

// Exists checks whether a project with the name `projectName` exists and returns an error if any
func (o kubernetesClient) Exists(projectName string) (bool, error) {
	projectSupport, err := o.client.IsProjectSupported()
	if err != nil {
		return false, fmt.Errorf("unable to detect project support: %w", err)
	}

	if projectSupport {
		project, err := o.client.GetProject(projectName)
		if err != nil || project == nil {
			return false, err
		}
	} else {
		namespace, err := o.client.GetNamespace(projectName)
		if err != nil || namespace == nil {
			return false, err
		}
	}

	return true, nil
}
