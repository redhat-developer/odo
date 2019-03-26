package application

import (
	"os"
	"reflect"
	"regexp"
	"testing"

	appsv1 "github.com/openshift/api/apps/v1"
	applabels "github.com/openshift/odo/pkg/application/labels"
	"github.com/openshift/odo/pkg/component"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/preference"
	"github.com/openshift/odo/pkg/testingutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestGetDefaultAppName(t *testing.T) {
	tests := []struct {
		testName   string
		wantRE     string
		needPrefix bool
		prefix     string
	}{
		{
			testName:   "Case: App prefix not configured",
			wantRE:     "app-*",
			needPrefix: false,
		},
		{
			testName:   "Case: App prefix set to testing",
			wantRE:     "testing-*",
			needPrefix: true,
			prefix:     "testing",
		},
		{
			testName:   "Case: App prefix set to empty string",
			wantRE:     "application-*",
			needPrefix: true,
			prefix:     "",
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {

			odoConfigFile, kubeConfigFile, err := testingutil.SetUp(
				testingutil.ConfigDetails{
					FileName:      "odo-test-config",
					Config:        testingutil.FakeOdoConfig("odo-test-config", tt.needPrefix, tt.prefix),
					ConfigPathEnv: "GLOBALODOCONFIG",
				}, testingutil.ConfigDetails{
					FileName:      "kube-test-config",
					Config:        testingutil.FakeKubeClientConfig(),
					ConfigPathEnv: "KUBECONFIG",
				},
			)
			defer testingutil.CleanupEnv([]*os.File{odoConfigFile, kubeConfigFile}, t)
			if err != nil {
				t.Errorf("failed setting up the test env. Error: %v", err)
			}

			name, err := GetDefaultAppName()
			if err != nil {
				t.Errorf("failed to setup mock environment. Error: %v", err)
			}

			r, _ := regexp.Compile(tt.wantRE)
			match := r.MatchString(name)
			if !match {
				fetchedConfig, _ := preference.New()
				t.Errorf("randomly generated application name %s does not match regexp %s and config is %+v\nthe prefix is %s", name, tt.wantRE, fetchedConfig, *fetchedConfig.OdoSettings.NamePrefix)
			}
		})
	}
}

func TestGetMachineReadableFormat(t *testing.T) {
	type args struct {
		// client      *occlient.Client
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

	dcList := appsv1.DeploymentConfigList{
		Items: []appsv1.DeploymentConfig{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "frontend-myapp",
					Namespace: "myproject",
					Labels: map[string]string{
						applabels.ApplicationLabel:         "myapp",
						componentlabels.ComponentLabel:     "frontend",
						componentlabels.ComponentTypeLabel: "nodejs",
					},
					Annotations: map[string]string{
						component.ComponentSourceTypeAnnotation: "local",
					},
				},
				Spec: appsv1.DeploymentConfigSpec{
					Template: &corev1.PodTemplateSpec{
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
					},
					Annotations: map[string]string{
						component.ComponentSourceTypeAnnotation: "local",
					},
				},
				Spec: appsv1.DeploymentConfigSpec{
					Template: &corev1.PodTemplateSpec{
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
			client, fakeClientSet := occlient.FakeNew()
			//fake the dcs
			fakeClientSet.AppsClientset.PrependReactor("list", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &dcList, nil
			})

			for i := range dcList.Items {
				fakeClientSet.AppsClientset.PrependReactor("get", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
					return true, &dcList.Items[i], nil
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
						Status: AppStatus{
							Active: true,
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
						Status: AppStatus{
							Active: true,
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
