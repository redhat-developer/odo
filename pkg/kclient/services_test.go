package kclient

import (
	"reflect"
	"testing"

	"github.com/devfile/library/pkg/devfile/parser/data"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/generator"
	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/devfile/library/pkg/testingutil"
	componentlabels "github.com/openshift/odo/v2/pkg/component/labels"
	odoTestingUtil "github.com/openshift/odo/v2/pkg/testingutil"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ktesting "k8s.io/client-go/testing"
)

func TestCreateService(t *testing.T) {

	devObj := devfileParser.DevfileObj{
		Data: func() data.DevfileData {
			devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
			if err != nil {
				t.Error(err)
			}
			err = devfileData.AddComponents([]devfilev1.Component{testingutil.GetFakeContainerComponent("container1")})
			if err != nil {
				t.Error(err)
			}
			return devfileData
		}(),
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

			serviceParams := generator.ServiceParams{
				ObjectMeta:     objectMeta,
				SelectorLabels: labels,
			}

			service, err := generator.GetService(devObj, serviceParams, parsercommon.DevfileOptions{})
			if err != nil {
				t.Errorf("generator.GetService unexpected error %v", err)
			}

			createdService, err := fkclient.CreateService(*service)

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

	devObj := devfileParser.DevfileObj{
		Data: func() data.DevfileData {
			devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
			if err != nil {
				t.Error(err)
			}
			err = devfileData.AddComponents([]devfilev1.Component{testingutil.GetFakeContainerComponent("container1")})
			if err != nil {
				t.Error(err)
			}
			return devfileData
		}(),
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

			serviceParams := generator.ServiceParams{
				ObjectMeta:     objectMeta,
				SelectorLabels: labels,
			}

			service, err := generator.GetService(devObj, serviceParams, parsercommon.DevfileOptions{})
			if err != nil {
				t.Errorf("generator.GetService unexpected error %v", err)
			}

			updatedService, err := fkclient.UpdateService(*service)

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

func TestListServices(t *testing.T) {
	type args struct {
		selector string
	}
	tests := []struct {
		name             string
		args             args
		returnedServices corev1.ServiceList
		want             []corev1.Service
		wantErr          bool
	}{
		{
			name: "case 1: returned 3 services",
			args: args{
				selector: componentlabels.GetSelector("nodejs", "app"),
			},
			returnedServices: corev1.ServiceList{
				Items: odoTestingUtil.FakeKubeServices("nodejs"),
			},
			want: odoTestingUtil.FakeKubeServices("nodejs"),
		},
		{
			name: "case 2: no service returned",
			args: args{
				selector: componentlabels.GetSelector("nodejs", "app"),
			},
			returnedServices: corev1.ServiceList{
				Items: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = "default"

			fkclientset.Kubernetes.PrependReactor("list", "services", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &tt.returnedServices, nil
			})

			got, err := fkclient.ListServices(tt.args.selector)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListServices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListServices() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetOneServiceFromSelector(t *testing.T) {
	wantService := odoTestingUtil.FakeKubeService("nodejs", "nodejs-app")

	type args struct {
		selector string
	}
	tests := []struct {
		name             string
		args             args
		returnedServices corev1.ServiceList
		want             *corev1.Service
		wantErr          bool
	}{

		{
			name: "case 1: returned the correct service",
			args: args{
				selector: componentlabels.GetSelector("nodejs", "app"),
			},
			returnedServices: corev1.ServiceList{
				Items: []corev1.Service{
					odoTestingUtil.FakeKubeService("nodejs", "nodejs-app"),
				},
			},
			want: &wantService,
		},
		{
			name: "case 2: no service returned",
			args: args{
				selector: componentlabels.GetSelector("nodejs", "app"),
			},
			returnedServices: corev1.ServiceList{
				Items: nil,
			},
			wantErr: true,
		},
		{
			name: "case 3: more than one service returned",
			args: args{
				selector: componentlabels.GetSelector("nodejs", "app"),
			},
			returnedServices: corev1.ServiceList{
				Items: odoTestingUtil.FakeKubeServices("nodejs"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = "default"

			fkclientset.Kubernetes.PrependReactor("list", "services", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &tt.returnedServices, nil
			})

			got, err := fkclient.GetOneServiceFromSelector(tt.args.selector)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOneServiceFromSelector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOneServiceFromSelector() got = %v, want %v", got, tt.want)
			}
		})
	}
}
