package project

import (
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"

	projectv1 "github.com/openshift/api/project/v1"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/testingutil"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery/fake"
	ktesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/clientcmd"
)

type resourceMapEntry struct {
	list *metav1.APIResourceList
	err  error
}

type fakeDiscovery struct {
	*fake.FakeDiscovery

	lock        sync.Mutex
	resourceMap map[string]*resourceMapEntry
}

var fakeDiscoveryWithProject = &fakeDiscovery{
	resourceMap: map[string]*resourceMapEntry{
		"project.openshift.io/v1": {
			list: &metav1.APIResourceList{
				GroupVersion: "project.openshift.io/v1",
				APIResources: []metav1.APIResource{{
					Name:         "projects",
					SingularName: "project",
					Namespaced:   false,
					Kind:         "Project",
					ShortNames:   []string{"proj"},
				}},
			},
		},
	},
}

var fakeDiscoveryWithNamespace = &fakeDiscovery{
	resourceMap: map[string]*resourceMapEntry{
		"v1": {
			list: &metav1.APIResourceList{
				GroupVersion: "v1",
				APIResources: []metav1.APIResource{{
					Name:         "namespaces",
					SingularName: "namespace",
					Namespaced:   false,
					Kind:         "Namespace",
					ShortNames:   []string{"ns"},
				}},
			},
		},
	},
}

func (c *fakeDiscovery) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if rl, ok := c.resourceMap[groupVersion]; ok {
		return rl.list, rl.err
	}
	return nil, kerrors.NewNotFound(schema.GroupResource{}, "")
}

func fakeDeleteKClient(namespace string, deleteLast bool) (*kclient.Client, *kclient.FakeClientset) {
	// Fake the client with the appropriate arguments
	client, fakeClientSet := kclient.FakeNew()

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	client.KubeConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	client.Namespace = "testing"
	fkWatch := watch.NewFake()

	fakeClientSet.Kubernetes.PrependReactor("list", "namespaces", func(action ktesting.Action) (bool, runtime.Object, error) {
		if deleteLast {
			return true, testingutil.FakeOnlyOneExistingNamespace(), nil
		}
		return true, testingutil.FakeProjects(), nil
	})

	fakeClientSet.Kubernetes.PrependReactor("delete", "namespaces", func(action ktesting.Action) (bool, runtime.Object, error) {
		return true, nil, nil
	})

	// We pass in the fakeNamespace in order to avoid race conditions with multiple go routines
	fakeNamespace := testingutil.FakeNamespaceStatus(corev1.NamespacePhase(""), namespace)
	go func(namespace *corev1.Namespace) {
		fkWatch.Delete(namespace)
	}(fakeNamespace)

	fakeClientSet.Kubernetes.PrependWatchReactor("namespaces", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
		return true, fkWatch, nil
	})

	return client, fakeClientSet
}

func fakeDeleteClient(namespace string, deleteLast bool) (*occlient.Client, *occlient.FakeClientset) {
	// Fake the client with the appropriate arguments
	client, fakeClientSet := occlient.FakeNew()

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	client.KubeConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	client.Namespace = "testing"
	fkWatch := watch.NewFake()

	fakeClientSet.ProjClientset.PrependReactor("list", "projects", func(action ktesting.Action) (bool, runtime.Object, error) {
		if deleteLast {
			return true, testingutil.FakeOnlyOneExistingProjects(), nil
		}
		return true, testingutil.FakeProjects(), nil
	})

	fakeClientSet.ProjClientset.PrependReactor("delete", "projects", func(action ktesting.Action) (bool, runtime.Object, error) {
		return true, nil, nil
	})

	// We pass in the fakeProject in order to avoid race conditions with multiple go routines
	fakeProject := testingutil.FakeProjectStatus(corev1.NamespacePhase(""), namespace)
	go func(project *projectv1.Project) {
		fkWatch.Delete(project)
	}(fakeProject)

	fakeClientSet.ProjClientset.PrependWatchReactor("projects", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
		return true, fkWatch, nil
	})

	return client, fakeClientSet
}

func fakeCreateKClient(namespace string) (*kclient.Client, *kclient.FakeClientset) {

	// Fake the client with the appropriate arguments
	client, fakeClientSet := kclient.FakeNew()

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	client.KubeConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	fkWatch := watch.NewFake()

	fakeClientSet.Kubernetes.PrependReactor("create", "namespace", func(action ktesting.Action) (bool, runtime.Object, error) {
		return true, nil, nil
	})

	fakeNamespace := testingutil.FakeNamespaceStatus(corev1.NamespacePhase("Active"), namespace)
	go func(project *corev1.Namespace) {
		fkWatch.Add(project)
	}(fakeNamespace)

	fakeClientSet.Kubernetes.PrependWatchReactor("namespaces", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
		return true, fkWatch, nil
	})

	kFkWatch2 := watch.NewFake()
	go func() {
		kFkWatch2.Add(&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
		})
	}()

	fakeClientSet.Kubernetes.PrependWatchReactor("serviceaccounts", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
		return true, kFkWatch2, nil
	})

	return client, fakeClientSet
}

func fakeCreateClient(namespace string) (*occlient.Client, *occlient.FakeClientset) {

	// Fake the client with the appropriate arguments
	client, fakeClientSet := occlient.FakeNew()

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	client.KubeConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	fkWatch := watch.NewFake()

	fakeClientSet.ProjClientset.PrependReactor("create", "projectrequests", func(action ktesting.Action) (bool, runtime.Object, error) {
		return true, nil, nil
	})

	fakeProject := testingutil.FakeProjectStatus(corev1.NamespacePhase("Active"), namespace)
	go func(project *projectv1.Project) {
		fkWatch.Add(project)
	}(fakeProject)

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

	return client, fakeClientSet
}

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

	// run tests for OpenShift (Project)
	for _, tt := range tests {
		t.Run(fmt.Sprintf(" %s with Project", tt.name), func(t *testing.T) {

			client, fakeClientSet := fakeCreateClient(tt.projectName)
			kubeClient, _ := fakeCreateKClient(tt.projectName)

			// The function we are testing
			context := genericclioptions.NewFakeContext(tt.projectName, "app", "cmp", client, kubeClient)
			context.Client.GetKubeClient().SetDiscoveryInterface(fakeDiscoveryWithProject)

			err := Create(context, tt.projectName, true)

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

	// run tests for Kubernetes (Namespace)
	for _, tt := range tests {
		t.Run(fmt.Sprintf(" %s with Namespace", tt.name), func(t *testing.T) {
			client, _ := fakeCreateClient(tt.projectName)
			kclient, fakeKubeClientSet := fakeCreateKClient(tt.projectName)

			// The function we are testing
			context := genericclioptions.NewFakeContext(tt.projectName, "app", "cmp", client, kclient)
			context.Client.GetKubeClient().SetDiscoveryInterface(fakeDiscoveryWithNamespace)

			err := Create(context, tt.projectName, true)

			// Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("project Create() unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && !tt.wantErr {
				if len(fakeKubeClientSet.Kubernetes.Actions()) != 2 {
					t.Errorf("expected 2 ProjClientSet.Actions() in Project Create, got: %v", len(fakeKubeClientSet.Kubernetes.Actions()))
				}
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
		deleteLast  bool
	}{
		{
			name:        "Case 1: Test project delete for multiple projects",
			wantErr:     false,
			wait:        false,
			projectName: "prj2",
			deleteLast:  false,
		},
		{
			name:        "Case 2: Test delete the only remaining project",
			wantErr:     false,
			wait:        false,
			projectName: "testing",
			deleteLast:  true,
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

	// run as on OpenShift (with Project)
	for _, tt := range tests {
		t.Run(fmt.Sprintf(" %s with Project", tt.name), func(t *testing.T) {

			client, fakeClientSet := fakeDeleteClient(tt.projectName, tt.deleteLast)
			kclient, _ := fakeDeleteKClient(tt.projectName, tt.deleteLast)

			context := genericclioptions.NewFakeContext(tt.projectName, "app", "cmp", client, kclient)
			context.Client.GetKubeClient().SetDiscoveryInterface(fakeDiscoveryWithProject)

			// The function we are testing
			err := Delete(context, tt.projectName, tt.wait)

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

	// run as on Kubernetes (with Namespace)
	for _, tt := range tests {
		t.Run(fmt.Sprintf(" %s with Namespace", tt.name), func(t *testing.T) {

			client, _ := fakeDeleteClient(tt.projectName, tt.deleteLast)
			kubeClient, fakeKClientSet := fakeDeleteKClient(tt.projectName, tt.deleteLast)

			context := genericclioptions.NewFakeContext(tt.projectName, "app", "cmp", client, kubeClient)
			context.Client.GetKubeClient().SetDiscoveryInterface(fakeDiscoveryWithNamespace)

			// The function we are testing
			err := Delete(context, tt.projectName, tt.wait)

			if err == nil && !tt.wantErr {
				if len(fakeKClientSet.Kubernetes.Actions()) != 1 {
					t.Errorf("expected 1 ProjClientSet.Actions() in Project Delete, got: %v", len(fakeKClientSet.Kubernetes.Actions()))
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
		returnedProjects *projectv1.ProjectList
		expectedProjects ProjectList
	}{
		{
			name:             "Case 1: Multiple projects returned",
			wantErr:          false,
			returnedProjects: testingutil.FakeProjects(),
			expectedProjects: NewProjectList([]Project{
				NewProject("testing", false),
				NewProject("prj1", false),
				NewProject("prj2", false),
			}),
		},
		{
			name:             "Case 2: Single project returned",
			wantErr:          false,
			returnedProjects: testingutil.FakeOnlyOneExistingProjects(),
			expectedProjects: NewProjectList([]Project{
				NewProject("testing", false),
			}),
		},
		{
			name:             "Case 3: No project returned",
			wantErr:          false,
			returnedProjects: &projectv1.ProjectList{},
			expectedProjects: NewProjectList([]Project{}),
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

			kubeClient, _ := kclient.FakeNew()

			client.GetKubeClient().SetDiscoveryInterface(fakeDiscoveryWithProject)

			context := genericclioptions.NewFakeContext("test", "app", "cmp", client, kubeClient)

			// The function we are testing
			projects, err := List(context)

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
