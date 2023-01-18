package kclient

import (
	"errors"
	"testing"
	
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/generator"
	devfileParser "github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	parsercommon "github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	"github.com/devfile/library/v2/pkg/testingutil"
	
	odoTestingUtil "github.com/redhat-developer/odo/pkg/testingutil"
	
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ktesting "k8s.io/client-go/testing"
)

// createFakeDeployment creates a fake deployment with the given pod name and labels
func createFakeDeployment(fkclient *Client, fkclientset *FakeClientset, podName string, labels map[string]string) (*appsv1.Deployment, error) {
	fakeUID := types.UID("12345")
	devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
	if err != nil {
		return nil, err
	}

	err = devfileData.AddComponents([]devfilev1.Component{
		testingutil.GetFakeContainerComponent("container1"),
	})
	if err != nil {
		return nil, err
	}

	devObj := devfileParser.DevfileObj{Data: devfileData}

	containers, err := generator.GetContainers(devObj, parsercommon.DevfileOptions{})
	if err != nil {
		return nil, err
	}

	objectMeta := generator.GetObjectMeta(podName, "default", labels, nil)

	deploymentParams := generator.DeploymentParams{
		ObjectMeta: objectMeta,
		Containers: containers,
	}
	deploy, _ := generator.GetDeployment(devObj, deploymentParams)
	fkclientset.Kubernetes.PrependReactor("patch", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
		if podName == "" {
			return true, nil, errors.New("deployment name is empty")
		}
		deployment := appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       DeploymentKind,
				APIVersion: DeploymentAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: podName,
				UID:  fakeUID,
			},
		}
		return true, &deployment, nil
	})

	createdDeployment, err := fkclient.ApplyDeployment(*deploy)
	if err != nil {
		return nil, err
	}
	return createdDeployment, nil
}

func TestCreateDeployment(t *testing.T) {

	labels := map[string]string{
		"app":       "app",
		"component": "frontend",
	}

	tests := []struct {
		name           string
		deploymentName string
		wantErr        bool
	}{
		{
			name:           "Case: Valid deployment name",
			deploymentName: "pod",
			wantErr:        false,
		},
		{
			name:           "Case: Invalid deployment name",
			deploymentName: "",
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = "default"

			createdDeployment, err := createFakeDeployment(fkclient, fkclientset, tt.deploymentName, labels)
			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("fkclient.CreateDeployment(pod) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {

				if len(fkclientset.Kubernetes.Actions()) != 2 {
					t.Errorf("expected 2 action in StartDeployment got %d: %v", len(fkclientset.Kubernetes.Actions()), fkclientset.Kubernetes.Actions())
				} else {
					if createdDeployment.Name != tt.deploymentName {
						t.Errorf("deployment name does not match the expected name, expected: %s, got %s", tt.deploymentName, createdDeployment.Name)
					}
				}

			}

		})
	}
}

func TestGetDeploymentByName(t *testing.T) {

	tests := []struct {
		name               string
		deploymentName     string
		wantDeploymentName string
		wantErr            bool
	}{
		{
			name:               "Case 1: Valid deployment name",
			deploymentName:     "mydeploy1",
			wantDeploymentName: "mydeploy1",
			wantErr:            false,
		},
		{
			name:               "Case 2: Invalid deployment name",
			deploymentName:     "mydeploy2",
			wantDeploymentName: "",
			wantErr:            false,
		},
		{
			name:           "Case 3: Error condition",
			deploymentName: "mydeploy1",
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = "default"

			fkclientset.Kubernetes.PrependReactor("get", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.deploymentName == "mydeploy2" {
					emptyDeployment := odoTestingUtil.CreateFakeDeployment("", false)
					return true, emptyDeployment, nil
				} else if tt.deploymentName == "mydeploy1" {
					deployment := odoTestingUtil.CreateFakeDeployment(tt.deploymentName, false)
					return true, deployment, nil
				} else {
					return true, nil, errors.New("deployment get error")
				}

			})

			deployment, err := fkclient.GetDeploymentByName(tt.deploymentName)
			if !tt.wantErr && err != nil {
				t.Errorf("TestGetDeploymentByName unexpected error: %v", err)
			} else if !tt.wantErr && deployment.GetName() != tt.wantDeploymentName {
				t.Errorf("TestGetDeploymentByName error: expected %v, got %v", tt.wantDeploymentName, deployment.GetName())
			}

		})
	}
}

func TestUpdateDeployment(t *testing.T) {

	labels := map[string]string{
		"app":       "app",
		"component": "frontend",
	}

	devObj := devfileParser.DevfileObj{
		Data: func() data.DevfileData {
			devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
			_ = devfileData.AddComponents([]devfilev1.Component{
				testingutil.GetFakeContainerComponent("container1"),
			})
			return devfileData
		}(),
	}

	containers, err := generator.GetContainers(devObj, parsercommon.DevfileOptions{})
	if err != nil {
		t.Errorf("generator.GetContainers unexpected error %v", err)
	}

	tests := []struct {
		name           string
		deploymentName string
		wantErr        bool
	}{
		{
			name:           "Case: Valid deployment name",
			deploymentName: "pod",
			wantErr:        false,
		},
		{
			name:           "Case: Invalid deployment name",
			deploymentName: "",
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = "default"

			objectMeta := generator.GetObjectMeta(tt.deploymentName, "default", labels, nil)

			deploymentParams := generator.DeploymentParams{
				ObjectMeta: objectMeta,
				Containers: containers,
			}
			deploy, _ := generator.GetDeployment(devObj, deploymentParams)
			fkclientset.Kubernetes.PrependReactor("patch", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.deploymentName == "" {
					return true, nil, errors.New("deployment name is empty")
				}
				deployment := appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       DeploymentKind,
						APIVersion: DeploymentAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: tt.deploymentName,
					},
				}
				return true, &deployment, nil
			})

			updatedDeployment, err := fkclient.ApplyDeployment(*deploy)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("fkclient.UpdateDeployment(pod) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {

				if len(fkclientset.Kubernetes.Actions()) != 2 {
					t.Errorf("expected 2 action in UpdateDeployment got %d: %v", len(fkclientset.Kubernetes.Actions()), fkclientset.Kubernetes.Actions())
				} else {
					if updatedDeployment.Name != tt.deploymentName {
						t.Errorf("deployment name does not match the expected name, expected: %s, got %s", tt.deploymentName, updatedDeployment.Name)
					}
				}

			}

		})
	}
}
