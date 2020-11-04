package kclient

import (
	"testing"

	"github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/kclient/generator"
	"github.com/openshift/odo/pkg/testingutil"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ktesting "k8s.io/client-go/testing"
)

func TestCreateService(t *testing.T) {

	devObj := parser.DevfileObj{
		Data: &testingutil.TestDevfileData{
			Components: []common.DevfileComponent{
				testingutil.GetFakeContainerComponent("container1"),
			},
		},
	}

	tests := []struct {
		name          string
		componentName string
		wantErr       bool
	}{
		{
			name:          "Case: Valid component name",
			componentName: "testComponent",
			wantErr:       false,
		},
		{
			name:          "Case: Invalid component name",
			componentName: "",
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = "default"

			objectMeta := generator.GetObjectMeta(tt.componentName, "default", nil, nil)

			labels := map[string]string{
				"component": tt.componentName,
			}

			fkclientset.Kubernetes.PrependReactor("create", "services", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.componentName == "" {
					return true, nil, errors.Errorf("component name is empty")
				}
				service := corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name: tt.componentName,
					},
				}
				return true, &service, nil
			})

			serviceSpec, err := generator.GetService(devObj, labels)
			if err != nil {
				t.Errorf("generator.GetService unexpected error %v", err)
			}

			createdService, err := fkclient.CreateService(objectMeta, *serviceSpec)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("fkclient.createService unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if len(fkclientset.Kubernetes.Actions()) != 1 {
					t.Errorf("expected 1 action, got: %v", fkclientset.Kubernetes.Actions())
				} else {
					if createdService.Name != tt.componentName {
						t.Errorf("component name does not match the expected name, expected: %s, got %s", tt.componentName, createdService.Name)
					}
				}
			}

		})
	}
}

func TestUpdateService(t *testing.T) {

	devObj := parser.DevfileObj{
		Data: &testingutil.TestDevfileData{
			Components: []common.DevfileComponent{
				testingutil.GetFakeContainerComponent("container1"),
			},
		},
	}

	tests := []struct {
		name          string
		componentName string
		wantErr       bool
	}{
		{
			name:          "Case: Valid component name",
			componentName: "testComponent",
			wantErr:       false,
		},
		{
			name:          "Case: Invalid component name",
			componentName: "",
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = "default"

			objectMeta := generator.GetObjectMeta(tt.componentName, "default", nil, nil)

			labels := map[string]string{
				"component": tt.componentName,
			}

			fkclientset.Kubernetes.PrependReactor("update", "services", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.componentName == "" {
					return true, nil, errors.Errorf("component name is empty")
				}
				service := corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name: tt.componentName,
					},
				}
				return true, &service, nil
			})

			serviceSpec, err := generator.GetService(devObj, labels)
			if err != nil {
				t.Errorf("generator.GetService unexpected error %v", err)
			}

			updatedService, err := fkclient.UpdateService(objectMeta, *serviceSpec)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("fkclient.UpdateService unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {

				if len(fkclientset.Kubernetes.Actions()) != 1 {
					t.Errorf("expected 1 action, got: %v", fkclientset.Kubernetes.Actions())
				} else {
					if updatedService.Name != tt.componentName {
						t.Errorf("service name does not match the expected name, expected: %s, got %s", tt.componentName, updatedService.Name)
					}
				}

			}

		})
	}
}
