package v2

import (
	"fmt"
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"reflect"
	"strings"
)

// GetProjects returns the Project Object parsed from devfile
func (d *DevfileV2) GetProjects(options common.DevfileOptions) ([]v1.Project, error) {

	if reflect.DeepEqual(options, common.DevfileOptions{}) {
		return d.Projects, nil
	}

	var projects []v1.Project
	for _, project := range d.Projects {
		// Filter Project Attributes
		filterIn, err := common.FilterDevfileObject(project.Attributes, options)
		if err != nil {
			return nil, err
		} else if !filterIn {
			continue
		}

		// Filter Project Source Type - Git, Zip, etc.
		projectSourceType, err := common.GetProjectSourceType(project.ProjectSource)
		if err != nil {
			return nil, err
		}
		if options.ProjectOptions.ProjectSourceType != "" && projectSourceType != options.ProjectOptions.ProjectSourceType {
			continue
		}
		if options.FilterByName == "" || project.Name == options.FilterByName {
			projects = append(projects, project)
		}
	}

	return projects, nil
}

// AddProjects adss the slice of Devfile projects to the Devfile's project list
// a project is considered as invalid if it is already defined
// project list passed in will be all processed, and returns a total error of all invalid projects
func (d *DevfileV2) AddProjects(projects []v1.Project) error {
	projectsMap := make(map[string]bool)
	var errorsList []string
	for _, project := range d.Projects {
		projectsMap[project.Name] = true
	}

	for _, project := range projects {
		if _, ok := projectsMap[project.Name]; !ok {
			d.Projects = append(d.Projects, project)
		} else {
			errorsList = append(errorsList, (&common.FieldAlreadyExistError{Name: project.Name, Field: "project"}).Error())
			continue
		}
	}
	if len(errorsList) > 0 {
		return fmt.Errorf("errors while adding projects:\n%s", strings.Join(errorsList, "\n"))
	}
	return nil
}

// UpdateProject updates the slice of Devfile projects parsed from the Devfile
// return an error if the project is not found
func (d *DevfileV2) UpdateProject(project v1.Project) error {
	for i := range d.Projects {
		if d.Projects[i].Name == project.Name {
			d.Projects[i] = project
			return nil
		}
	}
	return fmt.Errorf("update project failed: project %s not found", project.Name)
}

// DeleteProject removes the specified project
func (d *DevfileV2) DeleteProject(name string) error {

	for i := range d.Projects {
		if d.Projects[i].Name == name {
			d.Projects = append(d.Projects[:i], d.Projects[i+1:]...)
			return nil
		}
	}

	return &common.FieldNotFoundError{
		Field: "project",
		Name:  name,
	}
}

//GetStarterProjects returns the DevfileStarterProject parsed from devfile
func (d *DevfileV2) GetStarterProjects(options common.DevfileOptions) ([]v1.StarterProject, error) {

	if reflect.DeepEqual(options, common.DevfileOptions{}) {
		return d.StarterProjects, nil
	}

	var starterProjects []v1.StarterProject
	for _, starterProject := range d.StarterProjects {
		// Filter Starter Project Attributes
		filterIn, err := common.FilterDevfileObject(starterProject.Attributes, options)
		if err != nil {
			return nil, err
		} else if !filterIn {
			continue
		}

		// Filter Starter Project Source Type - Git, Zip, etc.
		starterProjectSourceType, err := common.GetProjectSourceType(starterProject.ProjectSource)
		if err != nil {
			return nil, err
		}
		if options.ProjectOptions.ProjectSourceType != "" && starterProjectSourceType != options.ProjectOptions.ProjectSourceType {
			continue
		}

		if options.FilterByName == "" || starterProject.Name == options.FilterByName {
			starterProjects = append(starterProjects, starterProject)
		}
	}

	return starterProjects, nil
}

// AddStarterProjects adds the slice of Devfile starter projects to the Devfile's starter project list
// a starterProject is considered as invalid if it is already defined
// starterProject list passed in will be all processed, and returns a total error of all invalid starterProjects
func (d *DevfileV2) AddStarterProjects(projects []v1.StarterProject) error {
	projectsMap := make(map[string]bool)
	var errorsList []string
	for _, project := range d.StarterProjects {
		projectsMap[project.Name] = true
	}

	for _, project := range projects {
		if _, ok := projectsMap[project.Name]; !ok {
			d.StarterProjects = append(d.StarterProjects, project)
		} else {
			errorsList = append(errorsList, (&common.FieldAlreadyExistError{Name: project.Name, Field: "starterProject"}).Error())
			continue
		}
	}
	if len(errorsList) > 0 {
		return fmt.Errorf("errors while adding starterProjects:\n%s", strings.Join(errorsList, "\n"))
	}
	return nil
}

// UpdateStarterProject updates the slice of Devfile starter projects parsed from the Devfile
func (d *DevfileV2) UpdateStarterProject(project v1.StarterProject) error {
	for i := range d.StarterProjects {
		if d.StarterProjects[i].Name == project.Name {
			d.StarterProjects[i] = project
			return nil
		}
	}
	return fmt.Errorf("update starter project failed: starter project %s not found", project.Name)
}

// DeleteStarterProject removes the specified starter project
func (d *DevfileV2) DeleteStarterProject(name string) error {

	for i := range d.StarterProjects {
		if d.StarterProjects[i].Name == name {
			d.StarterProjects = append(d.StarterProjects[:i], d.StarterProjects[i+1:]...)
			return nil
		}
	}

	return &common.FieldNotFoundError{
		Field: "starter project",
		Name:  name,
	}
}
