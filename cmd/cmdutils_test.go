package cmd

import (
	"fmt"
	"reflect"
	"testing"

	appsv1 "github.com/openshift/api/apps/v1"
	"github.com/redhat-developer/odo/pkg/occlient"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func Test_validateName(t *testing.T) {
	tests := []struct {
		testCase string
		name     string
		wantErr  bool
	}{
		{
			testCase: "Test case - 1",
			name:     "app",
			wantErr:  false,
		},
		{
			testCase: "Test case - 2",
			name:     "app123",
			wantErr:  false,
		},
		{
			testCase: "Test case - 3",
			name:     "app-123",
			wantErr:  false,
		},
		{
			testCase: "Test case - 4",
			name:     "app.123",
			wantErr:  true,
		},
		{
			testCase: "Test case - 5",
			name:     "app_123",
			wantErr:  true,
		},
		{
			testCase: "Test case - 6",
			name:     "App",
			wantErr:  true,
		},
		{
			testCase: "Test case - 7",
			name:     "b7pdkva7ynxf8qoyuh02tbgu2ufcy4jq7udyom7it0g8gouc39x3gy0p1wrsbt6",
			wantErr:  false,
		},
		{
			testCase: "Test case - 8",
			name:     "b7pdkva7ynxf8qoyuh02tbgu2ufcy4jq7udyom7it0g8gouc39x3gy0p1wrsbt6p",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Log("Running test", tt.testCase)
		t.Run(tt.testCase, func(t *testing.T) {
			if err := validateName(tt.name); (err != nil) != tt.wantErr {
				t.Errorf("Expected error = %v, But got = %v", tt.wantErr, err)
			}
		})
	}
}

func Test_validateStoragePath(t *testing.T) {

	type args struct {
		storagePath, componentName, applicationName string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test Case 1",
			args: args{
				storagePath:     "/opt/app-root/src/storage/",
				componentName:   "nodejs",
				applicationName: "app",
			},
			wantErr: true,
		},

		{
			name: "Test Case 2",
			args: args{
				storagePath:     "/opt/app-root/src/storage/test",
				componentName:   "nodejs",
				applicationName: "app",
			},
			wantErr: false,
		},
	}

	pvcList := v1.PersistentVolumeClaimList{
		Items: []v1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "mystorage-app-pvc",
					Labels: map[string]string{
						"app.kubernetes.io/component-name": "nodejs",
						"app.kubernetes.io/name":           "app",
						"app.kubernetes.io/storage-name":   "mystorage",
					},
					Namespace: "myproject",
				},
			},
		},
	}

	pvc := v1.PersistentVolumeClaim{

		ObjectMeta: metav1.ObjectMeta{
			Name: "mystorage-app-pvc",
			Labels: map[string]string{
				"app.kubernetes.io/component-name": "nodejs",
				"app.kubernetes.io/name":           "app",
				"app.kubernetes.io/storage-name":   "mystorage",
			},
			Namespace: "myproject",
		},
	}

	listOfDC := appsv1.DeploymentConfigList{
		Items: []appsv1.DeploymentConfig{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nodejs-app",
					Namespace: "myproject",
					Labels: map[string]string{
						"app.kubernetes.io/component-name": "nodejs",
						"app.kubernetes.io/component-type": "nodejs",
						"app.kubernetes.io/name":           "app",
					},
				},
				Spec: appsv1.DeploymentConfigSpec{
					Template: &v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									VolumeMounts: []v1.VolumeMount{
										{
											MountPath: "/opt/app-root/src/storage/",
											Name:      "mystorage-app-pvc-idrcg-volume",
										},
									},
								},
							},

							Volumes: []v1.Volume{
								{
									Name: "mystorage-app-pvc-idrcg-volume",
									VolumeSource: v1.VolumeSource{
										PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
											ClaimName: "mystorage-app-pvc",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	labelSelector := "app.kubernetes.io/component-name=nodejs,app.kubernetes.io/name=app"
	storageSelector := "app.kubernetes.io/storage-name"
	client, fakeClientSet := occlient.FakeNew()
	fakeClientSet.AppsClientset.PrependReactor("list", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
		if !reflect.DeepEqual(action.(ktesting.ListAction).GetListRestrictions().Labels.String(), labelSelector) {
			return true, nil, fmt.Errorf("labels not matching with expected values, expected:%s, got:%s", labelSelector, action.(ktesting.ListAction).GetListRestrictions())
		}
		return true, &listOfDC, nil
	})

	fakeClientSet.Kubernetes.PrependReactor("get", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
		pvcName := action.(ktesting.GetAction).GetName()
		if pvcName != pvcList.Items[0].Name {
			return true, nil, fmt.Errorf("'get' called with different pvcName")
		}
		return true, &pvc, nil
	})

	fakeClientSet.Kubernetes.PrependReactor("list", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
		if !reflect.DeepEqual(action.(ktesting.ListAction).GetListRestrictions().Labels.String(), storageSelector) {
			return true, nil, fmt.Errorf("labels not matching with expected values, expected:%s, got:%s", storageSelector, action.(ktesting.ListAction).GetListRestrictions())
		}
		return true, &pvcList, nil
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := validateStoragePath(client, tt.args.storagePath, tt.args.componentName, tt.args.applicationName)
			if err != nil && tt.wantErr == false {
				t.Errorf("test failed, expected error: nil, but got: %#v", err)
			}

		})
	}

}
