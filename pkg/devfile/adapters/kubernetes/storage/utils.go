package storage

import (
	"fmt"

	"github.com/openshift/odo/pkg/kclient/generator"
	"github.com/openshift/odo/pkg/storage/labels"
	"github.com/pkg/errors"
	"k8s.io/klog"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/util"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const pvcNameMaxLen = 45

// CreateComponentStorage creates PVCs with the given list of storages if it does not exist, else it uses the existing PVC
func CreateComponentStorage(Client *kclient.Client, storages []common.Storage, componentName string) (err error) {

	for _, storage := range storages {
		volumeName := storage.Volume.Name
		volumeSize := storage.Volume.Size
		pvcName := storage.Name

		existingPVCName, err := GetExistingPVC(Client, volumeName, componentName)
		if err != nil {
			return err
		}

		if len(existingPVCName) == 0 {
			klog.V(2).Infof("Creating a PVC for %v", volumeName)
			_, err := Create(Client, volumeName, volumeSize, componentName, pvcName)
			if err != nil {
				return errors.Wrapf(err, "Error creating PVC for "+volumeName)
			}
		}
	}

	return
}

// Create creates the pvc for the given pvc name, volume name, volume size and component name
func Create(Client *kclient.Client, name, size, componentName, pvcName string) (*corev1.PersistentVolumeClaim, error) {

	labels := map[string]string{
		"component":                componentName,
		labels.DevfileStorageLabel: name,
	}

	quantity, err := resource.ParseQuantity(size)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse size: %v", size)
	}

	objectMeta := generator.GetObjectMeta(pvcName, Client.Namespace, labels, nil)
	pvcSpec := generator.GetPVCSpec(quantity)

	// Get the deployment
	deployment, err := Client.GetDeploymentByName(componentName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get deployment")
	}

	// Generate owner reference for the deployment and update objectMeta
	ownerReference := generator.GetOwnerReference(deployment)
	objectMeta.OwnerReferences = append(objectMeta.OwnerReferences, ownerReference)

	// Create PVC
	klog.V(2).Infof("Creating a PVC with name %v and labels %v", pvcName, labels)
	pvc, err := Client.CreatePVC(objectMeta, *pvcSpec)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create PVC")
	}
	return pvc, nil
}

// GeneratePVCNameFromDevfileVol generates a PVC name from the Devfile volume name and component name
func GeneratePVCNameFromDevfileVol(volName, componentName string) (string, error) {

	pvcName := fmt.Sprintf("%s-%s", volName, componentName)
	pvcName = util.TruncateString(pvcName, pvcNameMaxLen)
	randomChars := util.GenerateRandomString(4)
	pvcName, err := util.NamespaceOpenShiftObject(pvcName, randomChars)
	if err != nil {
		return "", errors.Wrapf(err, "unable to create namespaced name")
	}

	return pvcName, nil
}

// GetExistingPVC checks if a PVC is present and return the name if it exists
func GetExistingPVC(Client *kclient.Client, volumeName, componentName string) (string, error) {

	label := fmt.Sprintf("component=%s,%s=%s", componentName, labels.DevfileStorageLabel, volumeName)

	klog.V(2).Infof("Checking PVC for volume %v and label %v\n", volumeName, label)

	PVCs, err := Client.GetPVCsFromSelector(label)
	if err != nil {
		return "", errors.Wrapf(err, "Unable to get PVC with selectors "+label)
	}
	if len(PVCs) == 1 {
		klog.V(2).Infof("Found an existing PVC for volume %v and label %v\n", volumeName, label)
		existingPVC := &PVCs[0]
		return existingPVC.Name, nil
	} else if len(PVCs) == 0 {
		return "", nil
	} else {
		err = errors.New("More than 1 PVC found with the label " + label)
		return "", err
	}
}

// DeleteOldPVCs deletes all the old PVCs which are not in the processedVolumes map
func DeleteOldPVCs(Client *kclient.Client, componentName string, processedVolumes map[string]bool) error {
	label := fmt.Sprintf("component=%s", componentName)
	PVCs, err := Client.GetPVCsFromSelector(label)
	if err != nil {
		return errors.Wrapf(err, "unable to get PVC with selectors "+label)
	}
	for _, pvc := range PVCs {
		storageName, ok := pvc.GetLabels()[labels.DevfileStorageLabel]
		if ok && !processedVolumes[storageName] {
			// the pvc is not in the processedVolumes map
			// thus deleting those PVCs
			err := Client.DeletePVC(pvc.GetName())
			if err != nil {
				return err
			}
		}
	}
	return nil
}
