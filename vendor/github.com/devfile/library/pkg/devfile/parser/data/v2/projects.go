package v2

import (
	"strings"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

// GetProjects returns the Project Object parsed from devfile
func (d *DevfileV2) GetProjects(options common.DevfileOptions) ([]v1.Project, error) {
	if len(options.Filter) == 0 {
		return d.Projects, nil
	}

	var projects []v1.Project
	for _, proj := range d.Projects {
		filterIn, err := common.FilterDevfileObject(proj.Attributes, options)
		if err != nil {
			return nil, err
		}

		if filterIn {
			projects = append(projects, proj)
		}
	}

	return projects, nil
}

// AddProjects adss the slice of Devfile projects to the Devfile's project list
// if a project is already defined, error out
func (d *DevfileV2) AddProjects(projects []v1.Project) error {
	projectsMap := make(map[string]bool)
	for _, project := range d.Projects {
		projectsMap[project.Name] = true
	}

	for _, project := range projects {
		if _, ok := projectsMap[project.Name]; !ok {
			d.Projects = append(d.Projects, project)
		} else {
			return &common.FieldAlreadyExistError{Name: project.Name, Field: "project"}
		}
	}
	return nil
}

// UpdateProject updates the slice of Devfile projects parsed from the Devfile
func (d *DevfileV2) UpdateProject(project v1.Project) {
	for i := range d.Projects {
		if d.Projects[i].Name == strings.ToLower(project.Name) {
			d.Projects[i] = project
		}
	}
}

//GetStarterProjects returns the DevfileStarterProject parsed from devfile
func (d *DevfileV2) GetStarterProjects(options common.DevfileOptions) ([]v1.StarterProject, error) {
	if len(options.Filter) == 0 {
		return d.StarterProjects, nil
	}

	var starterProjects []v1.StarterProject
	for _, starterProj := range d.StarterProjects {
		filterIn, err := common.FilterDevfileObject(starterProj.Attributes, options)
		if err != nil {
			return nil, err
		}

		if filterIn {
			starterProjects = append(starterProjects, starterProj)
		}
	}

	return starterProjects, nil
}

// AddStarterProjects adds the slice of Devfile starter projects to the Devfile's starter project list
// if a starter project is already defined, error out
func (d *DevfileV2) AddStarterProjects(projects []v1.StarterProject) error {
	projectsMap := make(map[string]bool)
	for _, project := range d.StarterProjects {
		projectsMap[project.Name] = true
	}

	for _, project := range projects {
		if _, ok := projectsMap[project.Name]; !ok {
			d.StarterProjects = append(d.StarterProjects, project)
		} else {
			return &common.FieldAlreadyExistError{Name: project.Name, Field: "starterProject"}
		}
	}
	return nil
}

// UpdateStarterProject updates the slice of Devfile starter projects parsed from the Devfile
func (d *DevfileV2) UpdateStarterProject(project v1.StarterProject) {
	for i := range d.StarterProjects {
		if d.StarterProjects[i].Name == strings.ToLower(project.Name) {
			d.StarterProjects[i] = project
		}
	}
}
