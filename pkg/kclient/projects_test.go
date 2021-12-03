package kclient

import (
	"fmt"
	"reflect"
	"testing"

	projectv1 "github.com/openshift/api/project/v1"
	"github.com/redhat-developer/odo/pkg/testingutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	ktesting "k8s.io/client-go/testing"
)

func TestCreateNewProject(t *testing.T) {
	tests := []struct {
		name     string
		projName string
		wait     bool
		wantErr  bool
	}{
		{
			name:     "Case 1: valid project name, not waiting",
			projName: "testing",
			wait:     false,
			wantErr:  false,
		},
		{
			name:     "Case 2: valid project name, waiting",
			projName: "testing2",
			wait:     true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := FakeNew()

			fkclientset.ProjClientset.PrependReactor("create", "projectrequests", func(action ktesting.Action) (bool, runtime.Object, error) {
				proj := projectv1.Project{
					ObjectMeta: metav1.ObjectMeta{
						Name: tt.projName,
					},
				}
				return true, &proj, nil
			})

			if tt.wait {
				fkWatch := watch.NewFake()
				// Change the status
				go func() {
					fkWatch.Add(&projectv1.Project{
						ObjectMeta: metav1.ObjectMeta{
							Name: tt.projName,
						},
						Status: projectv1.ProjectStatus{Phase: "Active"},
					})
				}()
				fkclientset.ProjClientset.PrependWatchReactor("projects", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
					if len(tt.projName) == 0 {
						return true, nil, fmt.Errorf("error watching project")
					}
					return true, fkWatch, nil
				})
			}

			err := fkclient.CreateNewProject(tt.projName, tt.wait)
			if !tt.wantErr == (err != nil) {
				t.Errorf("client.CreateNewProject(string) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			actions := fkclientset.ProjClientset.Actions()
			actionsLen := len(actions)
			if !tt.wait && actionsLen != 1 {
				t.Errorf("expected 1 action in CreateNewProject got: %v", actions)
			}
			if tt.wait && actionsLen != 2 {
				t.Errorf("expected 2 actions in CreateNewProject when waiting for project creation got: %v", actions)
			}

			if err == nil {
				createdProj := actions[actionsLen-1].(ktesting.CreateAction).GetObject().(*projectv1.ProjectRequest)

				if createdProj.Name != tt.projName {
					t.Errorf("project name does not match the expected name, expected: %s, got: %s", tt.projName, createdProj.Name)
				}

				if tt.wait {
					expectedFields := fields.OneTermEqualSelector("metadata.name", tt.projName)
					gotFields := actions[0].(ktesting.WatchAction).GetWatchRestrictions().Fields

					if !reflect.DeepEqual(expectedFields, gotFields) {
						t.Errorf("Fields not matching: expected: %s, got %s", expectedFields, gotFields)
					}
				}
			}

		})
	}
}

func TestListProjects(t *testing.T) {
	tests := []struct {
		name             string
		returnedProjects *projectv1.ProjectList
		want             *projectv1.ProjectList
		wantErr          bool
	}{
		{
			name:             "case 1: three projects returned",
			returnedProjects: testingutil.FakeProjects(),
			want:             testingutil.FakeProjects(),
			wantErr:          false,
		},
		{
			name:             "case 2: no projects present",
			returnedProjects: &projectv1.ProjectList{},
			want:             &projectv1.ProjectList{},
			wantErr:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := FakeNew()

			fkclientset.ProjClientset.PrependReactor("list", "projects", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, tt.returnedProjects, nil
			})

			got, err := fkclient.ListProjects()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListProjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListProjects() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListProjectNames(t *testing.T) {
	tests := []struct {
		name             string
		returnedProjects *projectv1.ProjectList
		want             []string
		wantErr          bool
	}{
		{
			name:             "case 1: three projects returned",
			returnedProjects: testingutil.FakeProjects(),
			want:             []string{"testing", "prj1", "prj2"},
			wantErr:          false,
		},
		{
			name:             "case 2: no projects present",
			returnedProjects: &projectv1.ProjectList{},
			want:             nil,
			wantErr:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := FakeNew()

			fkclientset.ProjClientset.PrependReactor("list", "projects", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, tt.returnedProjects, nil
			})

			got, err := fkclient.ListProjectNames()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListProjectNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListProjectNames() got = %v, want %v", got, tt.want)
			}
		})
	}
}
