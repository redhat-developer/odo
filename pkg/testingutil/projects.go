package testingutil

import (
	projectv1 "github.com/openshift/api/project/v1"
	v1 "github.com/openshift/api/project/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getFakeProject(projectName string) v1.Project {
	return v1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: projectName,
		},
	}
}

// FakeProjects returns fake projectlist for use by API mock functions for Unit tests
func FakeProjects() *v1.ProjectList {
	return &v1.ProjectList{
		Items: []v1.Project{
			getFakeProject("testing"),
			getFakeProject("prj1"),
			getFakeProject("prj2"),
		},
	}
}

// FakeProjectStatus returns fake project status for use by mock watch on project
func FakeProjectStatus(prjStatus corev1.NamespacePhase, prjName string) *projectv1.Project {
	return &projectv1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: prjName,
		},
		Status: projectv1.ProjectStatus{Phase: prjStatus},
	}
}

// FakeOnlyOneExistingProjects returns fake projectlist with single project for use by API mock functions for Unit tests testing delete of the only available project
func FakeOnlyOneExistingProjects() *v1.ProjectList {
	return &v1.ProjectList{
		Items: []v1.Project{
			getFakeProject("testing"),
		},
	}
}

// FakeRemoveProject removes the delete requested project from the list of projects passed
func FakeRemoveProject(project string, projects *v1.ProjectList) *v1.ProjectList {
	for index, proj := range projects.Items {
		if proj.Name == project {
			projects.Items = append(projects.Items[:index], projects.Items[index+1:]...)
		}
	}
	return projects
}
