package storage

import (
	"fmt"

	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/preference"
	"github.com/openshift/odo/pkg/storage"
	storagelabels "github.com/openshift/odo/pkg/storage/labels"
	"github.com/openshift/odo/pkg/util"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GenerateVolumeNameFromPVC generates a volume name based on the pvc name
func GenerateVolumeNameFromPVC(pvc string) (volumeName string, err error) {
	volumeName, err = util.NamespaceOpenShiftObject(pvc, "vol")
	if err != nil {
		return "", err
	}
	return
}

// HandleEphemeralStorage creates or deletes the ephemeral volume based on the preference setting
func HandleEphemeralStorage(client kclient.Client, storageClient storage.Client, componentName string) error {
	pref, err := preference.New()
	if err != nil {
		return err
	}

	selector := fmt.Sprintf("%v=%s,%s=%s", componentlabels.ComponentLabel, componentName, storagelabels.SourcePVCLabel, storage.OdoSourceVolume)

	pvcs, err := client.ListPVCs(selector)
	if err != nil && !kerrors.IsNotFound(err) {
		return err
	}

	if !pref.GetEphemeralSourceVolume() {
		if len(pvcs) == 0 {
			err := storageClient.Create(storage.Storage{
				ObjectMeta: metav1.ObjectMeta{
					Name: storage.OdoSourceVolume,
				},
				Spec: storage.StorageSpec{
					Size: storage.OdoSourceVolumeSize,
				},
			})

			if err != nil {
				return err
			}
		} else if len(pvcs) > 1 {
			return fmt.Errorf("number of source volumes shouldn't be greater than 1")
		}
	} else {
		if len(pvcs) > 0 {
			for _, pvc := range pvcs {
				err := client.DeletePVC(pvc.Name)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
