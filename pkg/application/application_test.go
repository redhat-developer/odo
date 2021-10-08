package application

import (
	"reflect"
	"testing"

	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/testingutil"
	"github.com/openshift/odo/pkg/version"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//TODO: Fix this unit test. Mock occlient pkg with mockgen and use it for the test.
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
	//deploymentList := appsv1.DeploymentList{
	//	Items: []appsv1.Deployment{
	//		*testingutil.CreateFakeDeploymentsWithContainers("frontend", []corev1.Container{
	//			{Name: "dummyContainer"},
	//		}, []corev1.Container{}),
	//		*testingutil.CreateFakeDeploymentsWithContainers("backend", []corev1.Container{
	//			{Name: "dummyContainer"},
	//		}, []corev1.Container{}),
	//	},
	//}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fake the client with the appropriate arguments
			client, fakeClientSet := occlient.FakeNew()

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
			if got := GetMachineReadableFormat(client, tt.args.appName, tt.args.projectName); !reflect.DeepEqual(got, tt.want) {
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
			if got := GetMachineReadableFormatForList(tt.args.apps); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMachineReadableFormatForList() = %v, want %v", got, tt.want)
			}
		})
	}
}
