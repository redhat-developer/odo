package storage

import (
	"strings"
	"testing"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	v1 "k8s.io/api/core/v1"
)

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
