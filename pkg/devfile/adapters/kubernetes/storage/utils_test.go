package storage

import (
	"strings"
	"testing"

	"github.com/openshift/odo/pkg/kclient"
)

func TestStorageCreate(t *testing.T) {

	testComponentName := "test"

	tests := []struct {
		name        string
		volumeNames []string
	}{
		{
			name:        "storage test",
			volumeNames: []string{"vol1", "vol2", "vol3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, _ := kclient.FakeNew()

			// Create one of the test volumes
			createdPVC, err := Create(fkclient, tt.volumeNames[0], testComponentName)
			if err != nil {
				t.Errorf("Error creating PVC %v: %v", tt.volumeNames[0], err)
			}

			// It should create the remaining PVC and reuse the existing PVC
			volumeNameToPVC, err := CreateComponentStorage(fkclient, tt.volumeNames, testComponentName)
			if err != nil {
				t.Errorf("Error creating component storage %v: %v", tt.volumeNames, err)
			}

			if len(volumeNameToPVC) != len(tt.volumeNames) {
				t.Errorf("Incorrect number of volumes created, expected: %v actual: %v", len(tt.volumeNames), len(volumeNameToPVC))
			}

			volumeMatched := 0
			volumeReused := false
			for _, PVC := range volumeNameToPVC {
				for _, testVolumeName := range tt.volumeNames {
					if strings.Contains(PVC.Name, testVolumeName) {
						volumeMatched++
					}
				}
				if PVC.Name == createdPVC.Name {
					volumeReused = true
				}
			}

			if !volumeReused {
				t.Errorf("Volume not reused: %v", createdPVC.Name)
			}

			if volumeMatched != len(tt.volumeNames) {
				t.Errorf("Incorrect number of volumes matched, expected: %v actual: %v", len(tt.volumeNames), volumeMatched)
			}
		})
	}

}
