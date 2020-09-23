package testingutil

import (
	projectv1 "github.com/openshift/api/project/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getFakeProject(projectName string) projectv1.Project {
	return projectv1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: projectName,
		},
	}
}

func getFakeNamespace(name string) corev1.Namespace {
	return corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// FakeProjects returns fake projectlist for use by API mock functions for Unit tests
func FakeProjects() *projectv1.ProjectList {
	return &projectv1.ProjectList{
		Items: []projectv1.Project{
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

// FakeNamespaceStatus returns fake namespace status for use by mock watch on namespace
func FakeNamespaceStatus(status corev1.NamespacePhase, name string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Status: corev1.NamespaceStatus{Phase: status},
	}
}

// FakeOnlyOneExistingProjects returns fake projectlist with single project for use by API mock functions for Unit tests testing delete of the only available project
func FakeOnlyOneExistingProjects() *projectv1.ProjectList {
	return &projectv1.ProjectList{
		Items: []projectv1.Project{
			getFakeProject("testing"),
		},
	}
}

// FakeOnlyOneExistingNamespace similar as FakeOnlyOneExistingProjects only with Namespace
func FakeOnlyOneExistingNamespace() *corev1.NamespaceList {
	return &corev1.NamespaceList{
		Items: []corev1.Namespace{
			getFakeNamespace("testing"),
		},
	}
}
