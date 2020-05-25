package project

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"reflect"
	"testing"

	projectv1 "github.com/openshift/api/project/v1"
	v1 "github.com/openshift/api/project/v1"

	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/testingutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	ktesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/clientcmd"
)

func TestCreate(t *testing.T) {
	tests := []struct {
		name        string
		wantErr     bool
		projectName string
	}{
		{
			name:        "Case 1: project name is given",
			wantErr:     false,
			projectName: "project1",
		},
		{
			name:        "Case 2: no project name given",
			wantErr:     true,
			projectName: "",
		},
	}

	odoConfigFile, kubeConfigFile, err := testingutil.SetUp(
		testingutil.ConfigDetails{
			FileName:      "odo-test-config",
			Config:        testingutil.FakeOdoConfig("odo-test-config", false, ""),
			ConfigPathEnv: "GLOBALODOCONFIG",
		}, testingutil.ConfigDetails{
			FileName:      "kube-test-config",
			Config:        testingutil.FakeKubeClientConfig(),
			ConfigPathEnv: "KUBECONFIG",
		},
	)
	defer testingutil.CleanupEnv([]*os.File{odoConfigFile, kubeConfigFile}, t)
	if err != nil {
		t.Errorf("failed to create mock odo and kube config files. Error %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Fake the client with the appropriate arguments
			client, fakeClientSet := occlient.FakeNew()

			loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
			configOverrides := &clientcmd.ConfigOverrides{}
			client.KubeConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
			fkWatch := watch.NewFake()

			fakeClientSet.ProjClientset.PrependReactor("create", "projectrequests", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

			go func() {
				fkWatch.Add(testingutil.FakeProjectStatus(corev1.NamespacePhase("Active"), tt.projectName))
			}()
			fakeClientSet.ProjClientset.PrependWatchReactor("projects", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fkWatch, nil
			})

			fkWatch2 := watch.NewFake()
			go func() {
				fkWatch2.Add(&corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name: "default",
					},
				})
			}()

			fakeClientSet.Kubernetes.PrependWatchReactor("serviceaccounts", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fkWatch2, nil
			})

			// The function we are testing
			err := Create(client, tt.projectName, true)

			if err == nil && !tt.wantErr {
				if len(fakeClientSet.ProjClientset.Actions()) != 2 {
					t.Errorf("expected 2 ProjClientSet.Actions() in Project Create, got: %v", len(fakeClientSet.ProjClientset.Actions()))
				}
			}

			// Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("project Create() unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func TestDelete(t *testing.T) {
	tests := []struct {
		name        string
		wantErr     bool
		wait        bool
		projectName string
	}{
		{
			name:        "Case 1: Test project delete for multiple projects",
			wantErr:     false,
			wait:        false,
			projectName: "prj2",
		},
		{
			name:        "Case 2: Test delete the only remaining project",
			wantErr:     false,
			wait:        false,
			projectName: "testing",
		},
	}

	odoConfigFile, kubeConfigFile, err := testingutil.SetUp(
		testingutil.ConfigDetails{
			FileName:      "odo-test-config",
			Config:        testingutil.FakeOdoConfig("odo-test-config", false, ""),
			ConfigPathEnv: "GLOBALODOCONFIG",
		}, testingutil.ConfigDetails{
			FileName:      "kube-test-config",
			Config:        testingutil.FakeKubeClientConfig(),
			ConfigPathEnv: "KUBECONFIG",
		},
	)
	defer testingutil.CleanupEnv([]*os.File{odoConfigFile, kubeConfigFile}, t)
	if err != nil {
		t.Errorf("failed to create mock odo and kube config files. Error %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Fake the client with the appropriate arguments
			client, fakeClientSet := occlient.FakeNew()

			loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
			configOverrides := &clientcmd.ConfigOverrides{}
			client.KubeConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

			client.Namespace = "testing"
			fkWatch := watch.NewFake()

			fakeClientSet.ProjClientset.PrependReactor("list", "projects", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.name == "Test delete the only remaining project" {
					return true, testingutil.FakeOnlyOneExistingProjects(), nil
				}
				return true, testingutil.FakeProjects(), nil
			})

			fakeClientSet.ProjClientset.PrependReactor("delete", "projects", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

			// We pass in the fakeProject in order to avoid race conditions with multiple go routines
			fakeProject := testingutil.FakeProjectStatus(corev1.NamespacePhase(""), tt.projectName)
			go func(project *projectv1.Project) {
				fkWatch.Delete(project)
			}(fakeProject)

			fakeClientSet.ProjClientset.PrependWatchReactor("projects", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fkWatch, nil
			})

			// The function we are testing
			err := Delete(client, tt.projectName, tt.wait)

			if err == nil && !tt.wantErr {
				if len(fakeClientSet.ProjClientset.Actions()) != 1 {
					t.Errorf("expected 1 ProjClientSet.Actions() in Project Delete, got: %v", len(fakeClientSet.ProjClientset.Actions()))
				}
			}

			// Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("project Delete() unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		name             string
		wantErr          bool
		returnedProjects *v1.ProjectList
		expectedProjects ProjectList
	}{
		{
			name:             "Case 1: Multiple projects returned",
			wantErr:          false,
			returnedProjects: testingutil.FakeProjects(),
			expectedProjects: getMachineReadableFormatForList(
				[]Project{
					GetMachineReadableFormat("testing", false),
					GetMachineReadableFormat("prj1", false),
					GetMachineReadableFormat("prj2", false),
				},
			),
		},
		{
			name:             "Case 2: Single project returned",
			wantErr:          false,
			returnedProjects: testingutil.FakeOnlyOneExistingProjects(),
			expectedProjects: getMachineReadableFormatForList(
				[]Project{
					GetMachineReadableFormat("testing", false),
				},
			),
		},
		{
			name:             "Case 3: No project returned",
			wantErr:          false,
			returnedProjects: &v1.ProjectList{},
			expectedProjects: getMachineReadableFormatForList(
				nil,
			),
		},
	}

	odoConfigFile, kubeConfigFile, err := testingutil.SetUp(
		testingutil.ConfigDetails{
			FileName:      "odo-test-config",
			Config:        testingutil.FakeOdoConfig("odo-test-config", false, ""),
			ConfigPathEnv: "GLOBALODOCONFIG",
		}, testingutil.ConfigDetails{
			FileName:      "kube-test-config",
			Config:        testingutil.FakeKubeClientConfig(),
			ConfigPathEnv: "KUBECONFIG",
		},
	)
	defer testingutil.CleanupEnv([]*os.File{odoConfigFile, kubeConfigFile}, t)
	if err != nil {
		t.Errorf("failed to create mock odo and kube config files. Error %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Fake the client with the appropriate arguments
			client, fakeClientSet := occlient.FakeNew()

			loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
			configOverrides := &clientcmd.ConfigOverrides{}
			client.KubeConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

			fakeClientSet.ProjClientset.PrependReactor("list", "projects", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, tt.returnedProjects, nil
			})

			// The function we are testing
			projects, err := List(client)

			if !reflect.DeepEqual(projects, tt.expectedProjects) {
				t.Errorf("Expected project output is not equal, expected: %v, actual: %v", tt.expectedProjects, projects)
			}

			if err == nil && !tt.wantErr {
				if len(fakeClientSet.ProjClientset.Actions()) != 1 {
					t.Errorf("expected 1 ProjClientSet.Actions() in Project List, got: %v", len(fakeClientSet.ProjClientset.Actions()))
				}
			}

			// Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("project List() unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
