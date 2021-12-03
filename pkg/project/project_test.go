package project

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/kclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreate(t *testing.T) {

	tests := []struct {
		name                  string
		projectName           string
		wait                  bool
		isProjectSupported    bool
		isProjectSupportedErr error
		expectedErr           bool
	}{
		{
			name:        "empty project name",
			projectName: "",
			expectedErr: true,
		},
		{
			name:                  "new project without project resource",
			projectName:           "new-project",
			wait:                  false,
			isProjectSupported:    false,
			isProjectSupportedErr: nil,
			expectedErr:           false,
		},
		{
			name:                  "new project with project resource",
			projectName:           "new-project",
			wait:                  false,
			isProjectSupported:    true,
			isProjectSupportedErr: nil,
			expectedErr:           false,
		},
		{
			name:                  "new project without project resource and wait",
			projectName:           "new-project",
			wait:                  true,
			isProjectSupported:    false,
			isProjectSupportedErr: nil,
			expectedErr:           false,
		},
		{
			name:                  "new project with project resource and wait",
			projectName:           "new-project",
			wait:                  true,
			isProjectSupported:    true,
			isProjectSupportedErr: nil,
			expectedErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kc := kclient.NewMockClientInterface(ctrl)

			if tt.expectedErr == false {
				kc.EXPECT().IsProjectSupported().Return(tt.isProjectSupported, tt.isProjectSupportedErr)
				if tt.isProjectSupported {
					kc.EXPECT().CreateNewProject(tt.projectName, tt.wait).Times(1)
				} else {
					kc.EXPECT().CreateNamespace(tt.projectName).Times(1)
				}
				if tt.wait {
					kc.EXPECT().WaitForServiceAccountInNamespace(tt.projectName, "default").Times(1)
				}
			}

			err := Create(kc, tt.projectName, tt.wait)

			if err != nil != tt.expectedErr {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name                  string
		projectName           string
		wait                  bool
		isProjectSupported    bool
		isProjectSupportedErr error
		expectedErr           bool
	}{
		{
			name:        "empty project name",
			projectName: "",
			expectedErr: true,
		},
		{
			name:                  "delete project without project resource",
			projectName:           "new-project",
			wait:                  false,
			isProjectSupported:    false,
			isProjectSupportedErr: nil,
			expectedErr:           false,
		},
		{
			name:                  "delete project with project resource",
			projectName:           "new-project",
			wait:                  false,
			isProjectSupported:    true,
			isProjectSupportedErr: nil,
			expectedErr:           false,
		},
		{
			name:                  "delete project without project resource and wait",
			projectName:           "new-project",
			wait:                  true,
			isProjectSupported:    false,
			isProjectSupportedErr: nil,
			expectedErr:           false,
		},
		{
			name:                  "delete project with project resource and wait",
			projectName:           "new-project",
			wait:                  true,
			isProjectSupported:    true,
			isProjectSupportedErr: nil,
			expectedErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kc := kclient.NewMockClientInterface(ctrl)

			if tt.expectedErr == false {
				kc.EXPECT().IsProjectSupported().Return(tt.isProjectSupported, tt.isProjectSupportedErr)
				if tt.isProjectSupported {
					kc.EXPECT().DeleteProject(tt.projectName, tt.wait).Times(1)
				} else {
					kc.EXPECT().DeleteNamespace(tt.projectName, tt.wait).Times(1)
				}
			}

			err := Delete(kc, tt.projectName, tt.wait)

			if err != nil != tt.expectedErr {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestList(t *testing.T) {

	expectedList := ProjectList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: "odo.dev/v1alpha1",
		},
		Items: []Project{
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Project",
					APIVersion: "odo.dev/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "project1",
				},
			},
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Project",
					APIVersion: "odo.dev/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "project2",
				},
			},
		},
	}

	tests := []struct {
		name                  string
		isProjectSupported    bool
		isProjectSupportedErr error
		listNames             []string
		expectedErr           bool
		expectedList          ProjectList
	}{
		{
			name:                  "list projects without project resource",
			isProjectSupported:    false,
			listNames:             []string{"project1", "project2"},
			isProjectSupportedErr: nil,
			expectedErr:           false,
			expectedList:          expectedList,
		},
		{
			name:                  "list projects with project resource",
			isProjectSupported:    true,
			listNames:             []string{"project1", "project2"},
			isProjectSupportedErr: nil,
			expectedErr:           false,
			expectedList:          expectedList,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kc := kclient.NewMockClientInterface(ctrl)

			kc.EXPECT().GetCurrentNamespace().Times(1)

			if tt.expectedErr == false {
				kc.EXPECT().IsProjectSupported().Return(tt.isProjectSupported, tt.isProjectSupportedErr)
				if tt.isProjectSupported {
					kc.EXPECT().ListProjectNames().Return(tt.listNames, nil).Times(1)
				} else {
					kc.EXPECT().GetNamespaces().Return(tt.listNames, nil).Times(1)
				}
			}

			list, err := List(kc)

			if err != nil != tt.expectedErr {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				return
			}

			if !reflect.DeepEqual(list, tt.expectedList) {
				t.Errorf("Expected value:\n%+v\ngot:\n%+v", tt.expectedList, list)
			}
		})
	}
}

func TestExists(t *testing.T) {

	tests := []struct {
		name                  string
		projectName           string
		isProjectSupported    bool
		isProjectSupportedErr error
		expectedErr           bool
	}{
		{
			name:                  "project without project resource",
			projectName:           "new-project",
			isProjectSupported:    false,
			isProjectSupportedErr: nil,
			expectedErr:           false,
		},
		{
			name:                  "project with project resource",
			projectName:           "new-project",
			isProjectSupported:    true,
			isProjectSupportedErr: nil,
			expectedErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kc := kclient.NewMockClientInterface(ctrl)

			if tt.expectedErr == false {
				kc.EXPECT().IsProjectSupported().Return(tt.isProjectSupported, tt.isProjectSupportedErr)
				if tt.isProjectSupported {
					kc.EXPECT().GetProject(tt.projectName).Times(1)
				} else {
					kc.EXPECT().GetNamespace(tt.projectName).Times(1)
				}
			}

			_, err := Exists(kc, tt.projectName)

			if err != nil != tt.expectedErr {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}
