package project

// ToDo(@anmolbabu) uncomment tests when we have a nicer and cleaner way to stub occlient.go#ModifyConfig

import (
	"testing"

	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/testingutil"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Fake the client with the appropriate arguments
			client, fakeClientSet := occlient.FakeNew()
			occlient.SetCurrentProject = func(project string, c *occlient.Client) error {
				return nil
			}

			fakeClientSet.ProjClientset.PrependReactor("list", "projects", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.name == "Test delete the only remaining project" {
					return true, testingutil.FakeOnlyOneExistingProjects(), nil
				}
				return true, testingutil.FakeProjects(), nil
			})

			fakeClientSet.ProjClientset.PrependReactor("delete", "projects", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

			// The function we are testing
			err := Delete(client, tt.projectName)

			// Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("project Delete() unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
