package occlient

import (
	"github.com/openshift/odo/pkg/testingutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"strings"
	"testing"

	appsv1 "github.com/openshift/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestCreatePVC(t *testing.T) {
	tests := []struct {
		name    string
		size    string
		labels  map[string]string
		wantErr bool
	}{
		{
			name: "storage 10Gi",
			size: "10Gi",
			labels: map[string]string{
				"name":      "mongodb",
				"namespace": "blog",
			},
			wantErr: false,
		},
		{
			name: "storage 1024",
			size: "1024",
			labels: map[string]string{
				"name":      "PostgreSQL",
				"namespace": "backend",
			},
			wantErr: false,
		},
		{
			name: "storage invalid size",
			size: "4#0",
			labels: map[string]string{
				"name":      "MySQL",
				"namespace": "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := FakeNew()

			_, err := fkclient.CreatePVC(tt.name, tt.size, tt.labels)

			// Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf(" client.CreatePVC(name, size, labels) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			// Check for validating actions performed
			if (len(fkclientset.Kubernetes.Actions()) != 1) && (tt.wantErr != true) {
				t.Errorf("expected 1 action in CreatePVC got: %v", fkclientset.RouteClientset.Actions())
			}
			// Checks for return values in positive cases
			if err == nil {
				createdPVC := fkclientset.Kubernetes.Actions()[0].(ktesting.CreateAction).GetObject().(*corev1.PersistentVolumeClaim)
				quantity, err := resource.ParseQuantity(tt.size)
				if err != nil {
					t.Errorf("failed to create quantity by calling resource.ParseQuantity(%v)", tt.size)
				}

				// created PVC should be labeled with labels passed to CreatePVC
				if !reflect.DeepEqual(createdPVC.Labels, tt.labels) {
					t.Errorf("labels in created route is not matching expected labels, expected: %v, got: %v", tt.labels, createdPVC.Labels)
				}
				// name, size of createdPVC should be matching to size, name passed to CreatePVC
				if !reflect.DeepEqual(createdPVC.Spec.Resources.Requests["storage"], quantity) {
					t.Errorf("size of PVC is not matching to expected size, expected: %v, got %v", quantity, createdPVC.Spec.Resources.Requests["storage"])
				}
				if createdPVC.Name != tt.name {
					t.Errorf("PVC name is not matching to expected name, expected: %s, got %s", tt.name, createdPVC.Name)
				}
			}
		})
	}
}

func TestDeletePVC(t *testing.T) {
	tests := []struct {
		name    string
		pvcName string
		wantErr bool
	}{
		{
			name:    "storage 10Gi",
			pvcName: "postgresql",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			fakeClientSet.Kubernetes.PrependReactor("delete", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

			err := fakeClient.DeletePVC(tt.pvcName)

			//Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf(" client.DeletePVC(name) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			// Check for validating actions performed
			if (len(fakeClientSet.Kubernetes.Actions()) != 1) && (tt.wantErr != true) {
				t.Errorf("expected 1 action in DeletePVC got: %v", fakeClientSet.Kubernetes.Actions())
			}

			// Check for value with which the function has called
			DeletedPVC := fakeClientSet.Kubernetes.Actions()[0].(ktesting.DeleteAction).GetName()
			if DeletedPVC != tt.pvcName {
				t.Errorf("Delete action is performed with wrong pvcName, expected: %s, got %s", tt.pvcName, DeletedPVC)

			}
		})
	}
}

func TestAddPVCToDeploymentConfig(t *testing.T) {
	type args struct {
		dc   *appsv1.DeploymentConfig
		pvc  string
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test case 1: valid dc",
			args: args{
				dc: &appsv1.DeploymentConfig{
					Spec: appsv1.DeploymentConfigSpec{
						Selector: map[string]string{
							"deploymentconfig": "nodejs-app",
						},
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "test",
										VolumeMounts: []corev1.VolumeMount{
											{
												MountPath: "/tmp",
												Name:      "test",
											},
										},
									},
								},
							},
						},
					},
				},
				pvc:  "test volume",
				path: "/mnt",
			},
			wantErr: false,
		},
		{
			name: "Test case 2: dc without Containers defined",
			args: args{
				dc: &appsv1.DeploymentConfig{
					Spec: appsv1.DeploymentConfigSpec{
						Selector: map[string]string{
							"deploymentconfig": "nodejs-app",
						},
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{},
						},
					},
				},
				pvc:  "test-voulme",
				path: "/mnt",
			},
			wantErr: true,
		},
		{
			name: "Test case 3: dc without Template defined",
			args: args{
				dc: &appsv1.DeploymentConfig{
					Spec: appsv1.DeploymentConfigSpec{
						Selector: map[string]string{
							"deploymentconfig": "nodejs-app",
						},
					},
				},
				pvc:  "test-voulme",
				path: "/mnt",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, _ := FakeNew()

			err := fakeClient.AddPVCToDeploymentConfig(tt.args.dc, tt.args.pvc, tt.args.path)

			// Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("Client.AddPVCToDeploymentConfig() unexpected error = %v, wantErr %v", err, tt.wantErr)
			}

			// Checks for number of actions performed in positive cases
			if err == nil {

				found := false // creating a flag
				// iterating over the VolumeMounts for finding the one specified during func call
				for bb := range tt.args.dc.Spec.Template.Spec.Containers[0].VolumeMounts {
					if tt.args.path == tt.args.dc.Spec.Template.Spec.Containers[0].VolumeMounts[bb].MountPath {
						found = true
						if !strings.Contains(tt.args.dc.Spec.Template.Spec.Containers[0].VolumeMounts[bb].Name, tt.args.pvc) {
							t.Errorf("pvc name not matching with the specified value got: %v, expected %v", tt.args.dc.Spec.Template.Spec.Containers[0].VolumeMounts[bb].Name, tt.args.pvc)
						}
					}
				}
				if found == false {
					t.Errorf("expected Volume mount path %v not found in VolumeMounts", tt.args.path)
				}

				found = false // resetting the flag
				// iterating over the volume claims to find the one specified during func call
				for bb := range tt.args.dc.Spec.Template.Spec.Volumes {
					if tt.args.pvc == tt.args.dc.Spec.Template.Spec.Volumes[bb].VolumeSource.PersistentVolumeClaim.ClaimName {
						found = true
						if !strings.Contains(tt.args.dc.Spec.Template.Spec.Volumes[bb].Name, tt.args.pvc) {
							t.Errorf("pvc name not matching in PersistentVolumeClaim, got: %v, expected %v", tt.args.dc.Spec.Template.Spec.Volumes[bb].Name, tt.args.pvc)
						}
					}
				}
				if found == false {
					t.Errorf("expected volume %s not found in DeploymentConfig.Spec.Template.Spec.Volumes", tt.args.pvc)
				}

			}

		})
	}
}

func Test_updateStorageOwnerReference(t *testing.T) {
	fakeDC := fakeDeploymentConfig("dcName-1", "nodejs", nil, nil, t)
	type args struct {
		pvc            *corev1.PersistentVolumeClaim
		ownerReference []v1.OwnerReference
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "case 1: valid pvc",
			args: args{
				pvc: testingutil.FakePVC("pvc-1", "1Gi", map[string]string{}),
				ownerReference: []v1.OwnerReference{
					GenerateOwnerReference(fakeDC),
				},
			},
			wantErr: false,
		},
		{
			name: "case 2: empty owner reference for pvc",
			args: args{
				pvc:            testingutil.FakePVC("pvc-1", "1Gi", map[string]string{}),
				ownerReference: []v1.OwnerReference{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			fakeClientSet.Kubernetes.PrependReactor("get", "persistentvolumeclaims", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				returnedPVC := *tt.args.pvc
				return true, &returnedPVC, nil
			})

			fakeClientSet.Kubernetes.PrependReactor("update", "persistentvolumeclaims", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				pvc := action.(ktesting.UpdateAction).GetObject().(*corev1.PersistentVolumeClaim)
				if pvc.OwnerReferences == nil || pvc.OwnerReferences[0].Name != fakeDC.Name {
					t.Errorf("owner reference not set for dc %s", tt.args.pvc.Name)
				}
				return true, pvc, nil
			})

			if err := updateStorageOwnerReference(fakeClient, tt.args.pvc, tt.args.ownerReference...); (err != nil) != tt.wantErr {
				t.Errorf("updateStorageOwnerReference() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
