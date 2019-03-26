package project

import (
	"os"
	"testing"

	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/testingutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	ktesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/clientcmd"
)

func TestDelete(t *testing.T) {
	tests := []struct {
		name        string
		wantErr     bool
		projectName string
	}{
		{
			name:        "Test project delete for multiple projects",
			wantErr:     false,
			projectName: "prj2",
		},
		{
			name:        "Test delete the only remaining project",
			wantErr:     false,
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

			go func() {
				fkWatch.Delete(testingutil.FakeProjectStatus(corev1.NamespacePhase(""), tt.projectName))
			}()
			fakeClientSet.ProjClientset.PrependWatchReactor("projects", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fkWatch, nil
			})

			// The function we are testing
			err := Delete(client, tt.projectName)

			if err == nil && !tt.wantErr {
				if len(fakeClientSet.ProjClientset.Actions()) != 2 {
					t.Errorf("expected 2 ProjClientSet.Actions() in Project Delete, got: %v", len(fakeClientSet.ProjClientset.Actions()))
				}
			}

			// Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("project Delete() unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
