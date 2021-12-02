package project

import (
	"github.com/redhat-developer/odo/pkg/machineoutput"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ProjectKind = "Project"

type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Status            Status `json:"status,omitempty"`
}

type Status struct {
	Active bool `json:"active"`
}

// NewProject creates and returns a new project instance
func NewProject(projectName string, isActive bool) Project {
	return Project{
		TypeMeta: metav1.TypeMeta{
			Kind:       ProjectKind,
			APIVersion: machineoutput.APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: projectName,
		},
		Status: Status{
			Active: isActive,
		},
	}
}

type ProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Project `json:"items"`
}

// NewProjectList returns an instance of a list containing the `items` projects
func NewProjectList(items []Project) ProjectList {
	return ProjectList{
		TypeMeta: metav1.TypeMeta{
			Kind:       machineoutput.ListKind,
			APIVersion: machineoutput.APIVersion,
		},
		Items: items,
	}
}
