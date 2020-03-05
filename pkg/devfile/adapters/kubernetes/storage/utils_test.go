package storage

import (
	"strings"
	"testing"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/kclient"
)

func TestStorageCreate(t *testing.T) {

	testComponentName := "test"
	volNames := [...]string{"vol1", "vol2", "vol3"}
	volSize := "5Gi"

	tests := []struct {
		name    string
		volumes []common.Volume
	}{
		{
			name: "storage test",
			volumes: []common.Volume{
				{
					Name: &volNames[0],
					Size: &volSize,
				},
				{
					Name: &volNames[1],
					Size: &volSize,
				},
				{
					Name: &volNames[2],
					Size: &volSize,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, _ := kclient.FakeNew()

			// Create one of the test volumes
			createdPVC, err := Create(fkclient, *tt.volumes[0].Name, *tt.volumes[0].Size, testComponentName)
			if err != nil {
				t.Errorf("Error creating PVC %v: %v", *tt.volumes[0].Name, err)
			}

			// It should create the remaining PVC and reuse the existing PVC
			volumeNameToPVCName, err := CreateComponentStorage(fkclient, tt.volumes, testComponentName)
			if err != nil {
				t.Errorf("Error creating component storage %v: %v", tt.volumes, err)
			}

			if len(volumeNameToPVCName) != len(tt.volumes) {
				t.Errorf("Incorrect number of volumes created, expected: %v actual: %v", len(tt.volumes), len(volumeNameToPVCName))
			}

			volumeMatched := 0
			volumeReused := false
			for _, PVC := range volumeNameToPVCName {
				for _, testVolume := range tt.volumes {
					testVolumeName := *testVolume.Name
					if strings.Contains(PVC, testVolumeName) {
						volumeMatched++
					}
				}
				if PVC == createdPVC.Name {
					volumeReused = true
				}
			}

			if !volumeReused {
				t.Errorf("Volume not reused: %v", createdPVC.Name)
			}

			if volumeMatched != len(tt.volumes) {
				t.Errorf("Incorrect number of volumes matched, expected: %v actual: %v", len(tt.volumes), volumeMatched)
			}
		})
	}

}
