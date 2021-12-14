package application

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	gomock "github.com/golang/mock/gomock"

	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/testingutil"
	"github.com/redhat-developer/odo/pkg/unions"
	"github.com/redhat-developer/odo/pkg/version"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestList(t *testing.T) {
	ctrl := gomock.NewController(t)
	kubclient := kclient.NewMockClientInterface(ctrl)
	kubclient.EXPECT().GetDeploymentLabelValues("app.kubernetes.io/part-of", "app.kubernetes.io/part-of").Return([]string{"app1", "app3", "app1", "app2"}, nil).AnyTimes()
	appClient := NewClient(kubclient)
	result, err := appClient.List()
	expected := []string{"app1", "app2", "app3"}
	if err != nil {
		t.Errorf("Expected nil error, got %s", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Got %v, expected %v", result, expected)
	}
}

func TestExists(t *testing.T) {

	tests := []struct {
		name   string
		search string
		result bool
		err    bool
	}{
		{
			name:   "not exists",
			search: "an-app",
			result: false,
			err:    false,
		},
		{
			name:   "exists",
			search: "app1",
			result: true,
			err:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kubclient := kclient.NewMockClientInterface(ctrl)
			kubclient.EXPECT().GetDeploymentLabelValues("app.kubernetes.io/part-of", "app.kubernetes.io/part-of").Return([]string{"app1", "app3", "app1", "app2"}, nil).AnyTimes()
			appClient := NewClient(kubclient)
			result, err := appClient.Exists(tt.search)
			if err != nil != tt.err {
				t.Errorf("Expected %v error, got %v", tt.err, err)
			}
			if result != tt.result {
				t.Errorf("Expected %v, got %v", tt.result, result)
			}
		})
	}

}

func TestDelete(t *testing.T) {

	tests := []struct {
		name         string
		deleteReturn error
		expectedErr  string
	}{
		{
			name:         "kubernetes delete works",
			deleteReturn: nil,
			expectedErr:  "",
		},
		{
			name:         "kubernetes delete fails",
			deleteReturn: errors.New("an error"),
			expectedErr:  "unable to delete application",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kubclient := kclient.NewMockClientInterface(ctrl)
			appClient := NewClient(kubclient)
			labels := map[string]string{
				"app.kubernetes.io/part-of": "an-app",
			}
			kubclient.EXPECT().Delete(labels, false).Return(tt.deleteReturn).Times(1)

			// kube Delete works
			err := appClient.Delete("an-app")

			if err == nil && tt.expectedErr != "" {
				t.Errorf("Expected %v, got no error", tt.expectedErr)
				return
			}
			if err != nil && tt.expectedErr == "" {
				t.Errorf("Expected no error, got %v", err.Error())
				return
			}
			if err != nil && tt.expectedErr != "" && !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err.Error())
				return
			}
			if err != nil {
				return
			}
		})
	}
}

func TestComponentList(t *testing.T) {
	ctrl := gomock.NewController(t)
	kubclient := kclient.NewMockClientInterface(ctrl)
	depList := []appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app.kubernetes.io/instance": "a-component",
					"app.kubernetes.io/part-of":  "an-app-name",
				},
				Annotations: map[string]string{
					"odo.dev/project-type": "nodejs",
				},
			},
		},
	}
	kubclient.EXPECT().GetDeploymentFromSelector("app=an-app-name,app.kubernetes.io/managed-by=odo,app.kubernetes.io/part-of=an-app-name").Return(depList, nil).AnyTimes()
	kubclient.EXPECT().GetCurrentNamespace().Return("my-namespace").AnyTimes()
	kubclient.EXPECT().GetOneDeployment("a-component", "an-app-name").Return(&depList[0], nil).AnyTimes()
	ingresses := &unions.KubernetesIngressList{
		Items: nil,
	}
	kubclient.EXPECT().ListIngresses("app.kubernetes.io/instance=a-component,app.kubernetes.io/part-of=an-app-name").Return(ingresses, nil).AnyTimes()
	kubclient.EXPECT().IsServiceBindingSupported().Return(false, nil).AnyTimes()
	kubclient.EXPECT().ListSecrets("app.kubernetes.io/instance=a-component,app.kubernetes.io/part-of=an-app-name").Return(nil, nil).AnyTimes()
	kubclient.EXPECT().ListServices("").Return(nil, nil).AnyTimes()
	appClient := NewClient(kubclient)

	result, err := appClient.ComponentList("an-app-name")
	if len(result) != 1 {
		t.Errorf("expected 1 component in list, got %d", len(result))
	}
	component := result[0]
	if component.Name != "a-component" {
		t.Errorf("Expected component name %q, got %q", "a-component", component.Name)
	}
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
}

func TestGetMachineReadableFormat(t *testing.T) {
	type args struct {
		appName     string
		projectName string
		active      bool
	}
	tests := []struct {
		name string
		args args
		want App
	}{
		{

			name: "Test Case: machine readable output for application",
			args: args{
				appName:     "myapp",
				projectName: "myproject",
				active:      true,
			},
			want: App{
				TypeMeta: metav1.TypeMeta{
					Kind:       appKind,
					APIVersion: appAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myapp",
					Namespace: "myproject",
				},
				Spec: AppSpec{
					Components: []string{"frontend"},
				},
			},
		},
	}
	deploymentList := appsv1.DeploymentList{
		Items: []appsv1.Deployment{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "frontend-myapp",
					Namespace: "myproject",
					Labels: map[string]string{
						applabels.ApplicationLabel:         "myapp",
						componentlabels.ComponentLabel:     "frontend",
						componentlabels.ComponentTypeLabel: "nodejs",
						applabels.ManagedBy:                "odo",
						applabels.ManagerVersion:           version.VERSION,
					},
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "dummyContainer",
								},
							},
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "backend-app",
					Namespace: "myproject",
					Labels: map[string]string{
						applabels.ApplicationLabel:         "app",
						componentlabels.ComponentLabel:     "backend",
						componentlabels.ComponentTypeLabel: "java",
						applabels.ManagedBy:                "odo",
						applabels.ManagerVersion:           version.VERSION,
					},
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "dummyContainer",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fake the client with the appropriate arguments
			client, fakeClientSet := kclient.FakeNew()

			// fake the project
			fakeClientSet.Kubernetes.PrependReactor("get", "projects", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &testingutil.FakeOnlyOneExistingProjects().Items[0], nil
			})

			//fake the deployments
			fakeClientSet.Kubernetes.PrependReactor("list", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &deploymentList, nil
			})

			for i := range deploymentList.Items {
				fakeClientSet.Kubernetes.PrependReactor("get", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
					return true, &deploymentList.Items[i], nil
				})
			}
			kclient := NewClient(client)
			if got := kclient.GetMachineReadableFormat(tt.args.appName, tt.args.projectName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMachineReadableFormat() = %v,\n want %v", got, tt.want)
			}
		})
	}
}

func TestGetMachineReadableFormatForList(t *testing.T) {
	type args struct {
		apps []App
	}
	tests := []struct {
		name string
		args args
		want AppList
	}{
		{
			name: "Test Case: Machine Readable for Application List",
			args: args{
				apps: []App{
					{
						TypeMeta: metav1.TypeMeta{
							Kind:       appKind,
							APIVersion: appAPIVersion,
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "myapp",
						},
						Spec: AppSpec{
							Components: []string{"frontend"},
						},
					},
				},
			},
			want: AppList{
				TypeMeta: metav1.TypeMeta{
					Kind:       appList,
					APIVersion: appAPIVersion,
				},
				ListMeta: metav1.ListMeta{},
				Items: []App{
					{
						TypeMeta: metav1.TypeMeta{
							Kind:       appKind,
							APIVersion: appAPIVersion,
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "myapp",
						},
						Spec: AppSpec{
							Components: []string{"frontend"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := kclient.FakeNew()
			kclient := NewClient(client)
			if got := kclient.GetMachineReadableFormatForList(tt.args.apps); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMachineReadableFormatForList() = %v, want %v", got, tt.want)
			}
		})
	}
}
