package testingutil

import (
	v1 "github.com/openshift/api/project/v1"
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
