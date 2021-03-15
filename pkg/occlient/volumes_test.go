package occlient

import (
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

func TestRemoveVolumeFromDC(t *testing.T) {
	type args struct {
		volName string
		dc      appsv1.DeploymentConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Case 1 - Test removing volumes",
			args: args{
				volName: "foo-s2idata",
				dc:      *fakeDeploymentConfig("foo", "bar", nil, nil, t),
			},
			wantErr: false,
		},
		{
			name: "Case 2 - Error out, test removing non-existant volume",
			args: args{
				volName: "doesnotexist",
				dc:      *fakeDeploymentConfig("foo", "bar", nil, nil, t),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := removeVolumeFromDC(tt.args.volName, &tt.args.dc)

			if tt.wantErr && err == nil {
				t.Errorf("Wanted an error, got a pass")
			}

			if err != nil && !tt.wantErr {
				t.Errorf("Got error: %s", err)
			}

			// Check that it was actually removed
			for _, j := range tt.args.dc.Spec.Template.Spec.Volumes {
				if j.Name == tt.args.volName {
					t.Errorf("volume %s still exists even after removeVolumeFromDC function, %+v", tt.args.volName, tt.args.dc.Spec.Template.Spec.Containers)
				}
			}

		})
	}
}

func Test_removeVolumeMountsFromDC(t *testing.T) {
	type args struct {
		volName string
		dc      appsv1.DeploymentConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Case 1 - Test removing volume mount",
			args: args{
				volName: "foo-s2idata",
				dc:      *fakeDeploymentConfig("foo", "bar", nil, nil, t),
			},
			wantErr: false,
		},
		{
			name: "Case 2 - Error out, test removing non-existant volume mount",
			args: args{
				volName: "doesnotexist",
				dc:      *fakeDeploymentConfig("foo", "bar", nil, nil, t),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := removeVolumeMountsFromDC(tt.args.volName, &tt.args.dc)

			if tt.wantErr && err == nil {
				t.Errorf("Wanted an error, got a pass")
			}

			if err != nil && !tt.wantErr {
				t.Errorf("Got error: %s", err)
			}

			// Check that it was actually removed
			for _, container := range tt.args.dc.Spec.Template.Spec.Containers {
				for _, volMount := range container.VolumeMounts {
					if volMount.Name == tt.args.volName {
						t.Errorf("volume mount %s still exists even after removeVolumeMountsFromDC function, %+v", tt.args.volName, tt.args.dc.Spec.Template.Spec.Containers)
					}
				}
			}

		})
	}
}

func TestGetPVCNameFromVolumeMountName(t *testing.T) {
	dcWithPVC := fakeDeploymentConfig("test", "test", nil, nil, nil)
	dcWithPVC.Spec.Template.Spec.Volumes = append(dcWithPVC.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: "test-pvc",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: "test-pvc",
			},
		},
	})

	type args struct {
		volumeMountName string
		dc              *appsv1.DeploymentConfig
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test case : Deployment config with given PVC",
			args: args{
				volumeMountName: "test-pvc",
				dc:              dcWithPVC,
			},
			want: "test-pvc",
		},
		{
			name: "Test case : Deployment config without given PVC",
			args: args{
				volumeMountName: "non-existent-pvc",
				dc:              fakeDeploymentConfig("test", "test", nil, nil, nil),
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, _ := FakeNew()

			returnValue := fakeClient.GetPVCNameFromVolumeMountName(tt.args.volumeMountName, tt.args.dc)

			// Check for validating return value
			if returnValue != tt.want {
				t.Errorf("error in return value got: %v, expected %v", returnValue, tt.want)
			}
		})
	}
}

func TestRemoveVolumeFromDeploymentConfig(t *testing.T) {
	type args struct {
		pvc    string
		dcName string
	}
	tests := []struct {
		name     string
		dcBefore *appsv1.DeploymentConfig
		args     args
		wantErr  bool
	}{
		{
			name: "Test case : 1",
			dcBefore: &appsv1.DeploymentConfig{
				Spec: appsv1.DeploymentConfigSpec{
					Selector: map[string]string{
						"deploymentconfig": "test",
					},
					Template: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "test",
									VolumeMounts: []corev1.VolumeMount{
										{
											MountPath: "/tmp",
											Name:      "test-pvc",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "test-pvc",
									VolumeSource: corev1.VolumeSource{
										PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
											ClaimName: "test-pvc",
										},
									},
								},
							},
						},
					},
				},
			},
			args: args{
				pvc:    "test-pvc",
				dcName: "test",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			fakeClientSet.AppsClientset.PrependReactor("get", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, tt.dcBefore, nil
			})
			fakeClientSet.AppsClientset.PrependReactor("update", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})
			err := fakeClient.RemoveVolumeFromDeploymentConfig(tt.args.pvc, tt.args.dcName)

			// Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf(" client.RemoveVolumeFromDeploymentConfig(pvc, dcName) unexpected error %v, wantErr %v", err, tt.wantErr)
			}
			// Check for validating number of actions performed
			if (len(fakeClientSet.AppsClientset.Actions()) != 2) && (tt.wantErr != true) {
				t.Errorf("expected 2 actions in GetPVCFromName got: %v", fakeClientSet.Kubernetes.Actions())
			}
			updatedDc := fakeClientSet.AppsClientset.Actions()[1].(ktesting.UpdateAction).GetObject().(*appsv1.DeploymentConfig)
			//	validating volume got removed from dc
			for _, volume := range updatedDc.Spec.Template.Spec.Volumes {
				if volume.PersistentVolumeClaim.ClaimName == tt.args.pvc {
					t.Errorf("expected volume with name : %v to be removed from dc", tt.args.pvc)
				}
			}
		})
	}
}
