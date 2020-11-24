package storage

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/testingutil"
	"github.com/pkg/errors"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ktesting "k8s.io/client-go/testing"
)

func TestCreateComponentStorage(t *testing.T) {

	testComponentName := "test"
	podName := "testpod"
	fakeUID := types.UID("12345")
	volNames := [...]string{"vol1", "vol2"}
	volSize := "5Gi"

	tests := []struct {
		name     string
		storages []common.Storage
	}{
		{
			name: "storage test",
			storages: []common.Storage{
				{
					Name: "vol1-pvc",
					Volume: common.DevfileVolume{
						Name: volNames[0],
						Size: volSize,
					},
				},
				{
					Name: "vol2-pvc",
					Volume: common.DevfileVolume{
						Name: volNames[1],
						Size: volSize,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := kclient.FakeNew()

			fkclientset.Kubernetes.PrependReactor("get", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				deployment := appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       kclient.DeploymentKind,
						APIVersion: kclient.DeploymentAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: podName,
						UID:  fakeUID,
					},
				}
				return true, &deployment, nil
			})

			// Create one of the test volumes
			createdPVC, err := Create(fkclient, tt.storages[0].Volume.Name, tt.storages[0].Volume.Size, testComponentName, tt.storages[0].Name)
			if err != nil {
				t.Errorf("Error creating PVC %v: %v", tt.storages[0].Name, err)
			}

			if createdPVC.Name != tt.storages[0].Name {
				t.Errorf("PVC created name mismatch, expected: %v actual: %v", tt.storages[0].Name, createdPVC.Name)
			}

			fkclientset.Kubernetes.PrependReactor("create", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				labels := map[string]string{
					"component":    testComponentName,
					"storage-name": tt.storages[1].Volume.Name,
				}
				PVC := testingutil.FakePVC(tt.storages[1].Name, tt.storages[1].Volume.Size, labels)
				return true, PVC, nil
			})

			// It should create the remaining PVC and reuse the existing PVC
			err = CreateComponentStorage(fkclient, tt.storages, testComponentName)
			if err != nil {
				t.Errorf("Error creating component storage %v: %v", tt.storages, err)
			}
		})
	}

}

func TestStorageCreate(t *testing.T) {

	testComponentName := "test"
	podName := "testpod"
	fakeUID := types.UID("12345")
	volNames := [...]string{"vol1", "vol2"}
	volSize := "5Gi"
	garbageVolSize := "abc"

	tests := []struct {
		name    string
		storage common.Storage
		wantErr bool
		err     error
	}{
		{
			name: "valid pvc",
			storage: common.Storage{
				Name: "vol1-pvc",
				Volume: common.DevfileVolume{
					Name: volNames[0],
					Size: volSize,
				},
			},
			wantErr: false,
			err:     nil,
		},
		{
			name: "pvc with long name",
			storage: common.Storage{
				Name: "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz",
				Volume: common.DevfileVolume{
					Name: volNames[0],
					Size: volSize,
				},
			},
			wantErr: true,
			err:     errors.New("Error creating PVC, name is greater than 63"),
		},
		{
			name: "pvc with no name",
			storage: common.Storage{
				Name: "",
				Volume: common.DevfileVolume{
					Name: volNames[0],
					Size: volSize,
				},
			},
			wantErr: true,
			err:     errors.New("Error creating PVC, name is empty"),
		},
		{
			name: "pvc with invalid size",
			storage: common.Storage{
				Name: "vol1-pvc",
				Volume: common.DevfileVolume{
					Name: volNames[0],
					Size: garbageVolSize,
				},
			},
			wantErr: true,
			err:     nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := kclient.FakeNew()

			fkclientset.Kubernetes.PrependReactor("get", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				deployment := appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       kclient.DeploymentKind,
						APIVersion: kclient.DeploymentAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: podName,
						UID:  fakeUID,
					},
				}
				return true, &deployment, nil
			})

			fkclientset.Kubernetes.PrependReactor("create", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				labels := map[string]string{
					"component":    testComponentName,
					"storage-name": tt.storage.Volume.Name,
				}
				if tt.wantErr {
					return true, nil, tt.err
				}
				PVC := testingutil.FakePVC(tt.storage.Name, tt.storage.Volume.Size, labels)
				return true, PVC, nil
			})

			// Create one of the test volumes
			createdPVC, err := Create(fkclient, tt.storage.Volume.Name, tt.storage.Volume.Size, testComponentName, tt.storage.Name)
			if !tt.wantErr && err != nil {
				t.Errorf("Error creating PVC %v: %v", tt.storage.Name, err)
			} else if tt.wantErr && err != nil {
				// don't perform further checks if we want an error
				return
			}

			if createdPVC.Name != tt.storage.Name {
				t.Errorf("PVC created name mismatch, expected: %v actual: %v", tt.storage.Name, createdPVC.Name)
			}
		})
	}

}

func TestDeleteOldPVCs(t *testing.T) {
	type args struct {
		componentName    string
		processedVolumes map[string]bool
	}
	tests := []struct {
		name            string
		args            args
		returnedPVCs    *v1.PersistentVolumeClaimList
		deletedPVCNames map[string]bool
		wantErr         bool
	}{
		{
			name: "case 1: delete the non processed PVCs",
			args: args{
				componentName: "nodejs",
				processedVolumes: map[string]bool{
					"pvc-0": true,
				},
			},
			returnedPVCs: &v1.PersistentVolumeClaimList{
				Items: []v1.PersistentVolumeClaim{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "pvc-1-random-string",
							Labels: map[string]string{
								"component":    "nodejs",
								"storage-name": "pvc-1",
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "pvc-0-random-string",
							Labels: map[string]string{
								"component":    "nodejs",
								"storage-name": "pvc-0",
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "pvc-3-random-string",
							Labels: map[string]string{
								"component":    "nodejs",
								"storage-name": "pvc-3",
							},
						},
					},
				},
			},
			deletedPVCNames: map[string]bool{"pvc-1-random-string": true, "pvc-3-random-string": true},
			wantErr:         false,
		},
		{
			name: "case 2: no PVC returned",
			args: args{
				componentName: "nodejs",
			},
			returnedPVCs: &v1.PersistentVolumeClaimList{
				Items: []v1.PersistentVolumeClaim{},
			},
			deletedPVCNames: map[string]bool{},
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		deletedVolumesMap := make(map[string]bool)

		fkClient, fkClientSet := kclient.FakeNew()

		fkClientSet.Kubernetes.PrependReactor("list", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
			return true, tt.returnedPVCs, nil
		})

		fkClientSet.Kubernetes.PrependReactor("delete", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
			pvcName := action.(ktesting.DeleteAction).GetName()
			if _, ok := tt.deletedPVCNames[pvcName]; !ok {
				return true, nil, fmt.Errorf("delete called on a processed volume: %s", pvcName)
			}
			deletedVolumesMap[pvcName] = true
			return true, nil, nil
		})

		t.Run(tt.name, func(t *testing.T) {
			err := DeleteOldPVCs(fkClient, tt.args.componentName, tt.args.processedVolumes)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteOldPVCs() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr == true && err != nil {
				return
			}

			if !reflect.DeepEqual(tt.deletedPVCNames, deletedVolumesMap) {
				t.Errorf("all volumes are not deleted, want: %v, got: %v", tt.deletedPVCNames, deletedVolumesMap)
			}
		})
	}
}

func TestGetPVC(t *testing.T) {

	tests := []struct {
		pvc        string
		volumeName string
	}{
		{
			pvc:        "mypvc",
			volumeName: "myvolume",
		},
	}

	for _, tt := range tests {
		t.Run(tt.volumeName, func(t *testing.T) {
			volume := getPVC(tt.volumeName, tt.pvc)

			if volume.Name != tt.volumeName {
				t.Errorf("TestGetPVC error: volume name does not match; expected %s got %s", tt.volumeName, volume.Name)
			}

			if volume.PersistentVolumeClaim.ClaimName != tt.pvc {
				t.Errorf("TestGetPVC error: pvc name does not match; expected %s got %s", tt.pvc, volume.PersistentVolumeClaim.ClaimName)
			}
		})
	}
}

func TestAddVolumeMountToPodTemplateSpec(t *testing.T) {

	tests := []struct {
		podName                string
		namespace              string
		serviceAccount         string
		pvc                    string
		volumeName             string
		containerMountPathsMap map[string][]string
		container              v1.Container
		labels                 map[string]string
		wantErr                bool
	}{
		{
			podName:        "podSpecTest",
			namespace:      "default",
			serviceAccount: "default",
			pvc:            "mypvc",
			volumeName:     "myvolume",
			containerMountPathsMap: map[string][]string{
				"container1": {"/tmp/path1", "/tmp/path2"},
			},
			container: v1.Container{
				Name:            "container1",
				Image:           "image1",
				ImagePullPolicy: v1.PullAlways,

				Command: []string{"tail"},
				Args:    []string{"-f", "/dev/null"},
				Env:     []v1.EnvVar{},
			},
			labels: map[string]string{
				"app":       "app",
				"component": "frontend",
			},
			wantErr: false,
		},
		{
			podName:        "podSpecTest",
			namespace:      "default",
			serviceAccount: "default",
			pvc:            "mypvc",
			volumeName:     "myvolume",
			containerMountPathsMap: map[string][]string{
				"container1": {"/tmp/path1", "/tmp/path2"},
			},
			container: v1.Container{},
			labels: map[string]string{
				"app":       "app",
				"component": "frontend",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.podName, func(t *testing.T) {
			containers := addVolumeMountToContainers([]v1.Container{tt.container}, tt.volumeName, tt.containerMountPathsMap)

			mountPathCount := 0
			for _, container := range containers {
				if container.Name == tt.container.Name {
					for _, volumeMount := range container.VolumeMounts {
						if volumeMount.Name == tt.volumeName {
							for _, mountPath := range tt.containerMountPathsMap[tt.container.Name] {
								if volumeMount.MountPath == mountPath {
									mountPathCount++
								}
							}
						}
					}
				}
			}

			if mountPathCount != len(tt.containerMountPathsMap[tt.container.Name]) {
				t.Errorf("Volume Mounts for %s have not been properly mounted to the podTemplateSpec", tt.volumeName)
			}
		})
	}
}

func TestGetPVCAndVolumeMount(t *testing.T) {

	volNames := [...]string{"volume1", "volume2", "volume3"}
	volContainerPath := [...]string{"/home/user/path1", "/home/user/path2", "/home/user/path3"}

	tests := []struct {
		name                    string
		podName                 string
		namespace               string
		labels                  map[string]string
		containers              []v1.Container
		volumeNameToPVCName     map[string]string
		componentAliasToVolumes map[string][]common.DevfileVolume
		wantErr                 bool
	}{
		{
			name:      "Case: Valid case",
			podName:   "podSpecTest",
			namespace: "default",
			labels: map[string]string{
				"app":       "app",
				"component": "frontend",
			},
			containers: []v1.Container{
				{
					Name: "container1",
				},
				{
					Name: "container2",
				},
			},
			volumeNameToPVCName: map[string]string{
				"volume1": "volume1-pvc",
				"volume2": "volume2-pvc",
				"volume3": "volume3-pvc",
			},
			componentAliasToVolumes: map[string][]common.DevfileVolume{
				"container1": []common.DevfileVolume{
					{
						Name:          volNames[0],
						ContainerPath: volContainerPath[0],
					},
					{
						Name:          volNames[0],
						ContainerPath: volContainerPath[1],
					},
					{
						Name:          volNames[1],
						ContainerPath: volContainerPath[2],
					},
				},
				"container2": []common.DevfileVolume{
					{
						Name:          volNames[1],
						ContainerPath: volContainerPath[1],
					},
					{
						Name:          volNames[2],
						ContainerPath: volContainerPath[2],
					},
				},
			},
			wantErr: false,
		},
		{
			name:      "Case: Error case",
			podName:   "podSpecTest",
			namespace: "default",
			labels: map[string]string{
				"app":       "app",
				"component": "frontend",
			},
			containers: []v1.Container{
				{
					Name: "container2",
				},
			},
			volumeNameToPVCName: map[string]string{
				"volume2": "",
				"volume3": "volume3-pvc",
			},
			componentAliasToVolumes: map[string][]common.DevfileVolume{
				"container2": []common.DevfileVolume{
					{
						Name:          volNames[1],
						ContainerPath: volContainerPath[1],
					},
					{
						Name:          volNames[2],
						ContainerPath: volContainerPath[2],
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			containers, pvcVols, err := GetPVCAndVolumeMount(tt.containers, tt.volumeNameToPVCName, tt.componentAliasToVolumes)
			if !tt.wantErr && err != nil {
				t.Errorf("TestGetPVCAndVolumeMount.AddPVCAndVolumeMount() unexpected error %v, wantErr %v", err, tt.wantErr)
			} else if tt.wantErr && err != nil {
				return
			} else if tt.wantErr && err == nil {
				t.Error("TestGetPVCAndVolumeMount.AddPVCAndVolumeMount() expected error but got nil")
				return
			}

			// The total number of expected volumes is equal to the number of volumes defined in the devfile
			expectedNumVolumes := len(tt.volumeNameToPVCName)

			// check the number of containers and volumes in the pod template spec
			if len(containers) != len(tt.containers) {
				t.Errorf("TestGetPVCAndVolumeMount error - Incorrect number of Containers found in the pod template spec, expected: %v found: %v", len(tt.containers), len(containers))
				return
			}
			if len(pvcVols) != expectedNumVolumes {
				t.Errorf("TestGetPVCAndVolumeMount error - incorrect amount of pvc volumes in pod template spec expected %v, actual %v", expectedNumVolumes, len(pvcVols))
				return
			}

			// check the volume mounts of the pod template spec containers
			for _, container := range containers {
				for testcontainerAlias, testContainerVolumes := range tt.componentAliasToVolumes {
					if container.Name == testcontainerAlias {
						// check if container has the correct number of volume mounts
						if len(container.VolumeMounts) != len(testContainerVolumes) {
							t.Errorf("TestGetPVCAndVolumeMount - Incorrect number of Volume Mounts found in the pod template spec container %v, expected: %v found: %v", container.Name, len(testContainerVolumes), len(container.VolumeMounts))
						}

						// check if container has the specified volume
						volumeMatched := 0
						for _, volumeMount := range container.VolumeMounts {
							for _, testVolume := range testContainerVolumes {
								testVolumeName := testVolume.Name
								testVolumePath := testVolume.ContainerPath
								if strings.Contains(volumeMount.Name, testVolumeName) && volumeMount.MountPath == testVolumePath {
									volumeMatched++
								}
							}
						}
						if volumeMatched != len(testContainerVolumes) {
							t.Errorf("TestGetPVCAndVolumeMount - Failed to match Volume Mounts for pod template spec container %v, expected: %v found: %v", container.Name, len(testContainerVolumes), volumeMatched)
						}
					}
				}
			}
		})
	}
}
