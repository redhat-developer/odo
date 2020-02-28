package storage

import (
	"strings"
	"testing"

	"github.com/openshift/odo/pkg/devfile"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes/component"
	versionsCommon "github.com/openshift/odo/pkg/devfile/versions/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/testingutil"
)

func TestStorageAdapter(t *testing.T) {

	testComponentName := "test"

	tests := []struct {
		name                       string
		componentType              versionsCommon.DevfileComponentType
		wantErr                    bool
		containerAliasToVolumeName map[string][]string
		volumeNameNameToMountPath  map[string]string
	}{
		{
			name:          "Case: Invalid devfile",
			componentType: "",
			wantErr:       true,
		},
		{
			name:          "Case: Valid devfile",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			wantErr:       false,
			containerAliasToVolumeName: map[string][]string{
				"alias1": []string{"myvolume1"},
				"alias2": []string{"myvolume1", "myvolume2"},
			},
			volumeNameNameToMountPath: map[string]string{
				"myvolume1": "/my/volume/mount/path1",
				"myvolume2": "/my/volume/mount/path2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfile.DevfileObj{
				Data: testingutil.TestDevfileData{
					ComponentType: tt.componentType,
				},
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: testComponentName,
				Devfile:       devObj,
			}

			numOfVolumes := 0
			numofComponents := 0
			for _, component := range devObj.Data.GetAliasedComponents() {
				numofComponents++
				if component.Volumes != nil {
					numOfVolumes++
				}
			}

			fkclient, _ := kclient.FakeNew()

			componentAdapter := component.New(adapterCtx, *fkclient)
			storageAdapter := New(adapterCtx, *fkclient)

			podTemplateSpec, err := componentAdapter.Initialize()

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component adapter initialize unexpected error %v, wantErr %v", err, tt.wantErr)
			} else if tt.wantErr && (err != nil) {
				// if we want an error, return since the remaining test is not valid
				return
			}

			err = storageAdapter.Start(podTemplateSpec)
			if !tt.wantErr == (err != nil) {
				t.Errorf("storage adapter start unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			// check the number of containers and volumes in the pod template spec
			if len(podTemplateSpec.Spec.Containers) != numofComponents {
				t.Errorf("Incorrect number of Containers found in the pod template spec, expected: %v found: %v", numofComponents, len(podTemplateSpec.Spec.Containers))
				return
			}
			if len(podTemplateSpec.Spec.Volumes) != numOfVolumes {
				t.Errorf("Incorrect number of Volumes found in the pod template spec, expected: %v found: %v", numOfVolumes, len(podTemplateSpec.Spec.Volumes))
				return
			}

			// check the volume mounts of the pod template spec containers
			for _, container := range podTemplateSpec.Spec.Containers {
				for testContainerAlias, testVolumeNames := range tt.containerAliasToVolumeName {
					if container.Name == testContainerAlias {
						// check if container has the correct number of volume mounts
						if len(container.VolumeMounts) != len(testVolumeNames) {
							t.Errorf("Incorrect number of Volume Mounts found in the pod template spec container %v, expected: %v found: %v", container.Name, len(testVolumeNames), len(container.VolumeMounts))
						}

						// check if container has the specified volume
						volumeMatched := 0
						for _, volumeMount := range container.VolumeMounts {
							for _, testVolumeName := range testVolumeNames {
								if strings.Contains(volumeMount.Name, testVolumeName) && volumeMount.MountPath == tt.volumeNameNameToMountPath[testVolumeName] {
									volumeMatched++
								}
							}
						}
						if volumeMatched != len(testVolumeNames) {
							t.Errorf("Failed to match Volume Mounts for pod template spec container %v, expected: %v found: %v", container.Name, len(testVolumeNames), volumeMatched)
						}
					}
				}
			}
		})
	}

}
