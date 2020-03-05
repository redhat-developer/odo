package storage

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/util"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// CreateComponentStorage creates PVCs with the given list of volume names if it does not exist, else it uses the existing PVC
func CreateComponentStorage(Client *kclient.Client, volumes []common.Volume, componentName string) (map[string]string, error) {
	volumeNameToPVCName := make(map[string]string)

	for _, vol := range volumes {
		volumeName := *vol.Name
		volumeSize := *vol.Size
		label := "component=" + componentName + ",storage-name=" + volumeName

		glog.V(3).Infof("Checking for PVC with name %v and label %v\n", volumeName, label)
		PVCs, err := Client.GetPVCsFromSelector(label)
		if err != nil {
			err = errors.New("Unable to get PVC with selectors " + label + ": " + err.Error())
			return nil, err
		}
		if len(PVCs) == 1 {
			glog.V(3).Infof("Found an existing PVC with name %v and label %v\n", volumeName, label)
			existingPVC := &PVCs[0]
			volumeNameToPVCName[volumeName] = existingPVC.Name
		} else if len(PVCs) == 0 {
			glog.V(3).Infof("Creating a PVC with name %v and label %v\n", volumeName, label)
			createdPVC, err := Create(Client, volumeName, volumeSize, componentName)
			volumeNameToPVCName[volumeName] = createdPVC.Name
			if err != nil {
				err = errors.New("Error creating PVC " + volumeName + ": " + err.Error())
				return nil, err
			}
		} else {
			err = errors.New("More than 1 PVC found with the label " + label + ": " + err.Error())
			return nil, err
		}
	}

	return volumeNameToPVCName, nil
}

// Create creates the pvc for the given pvc name and component name
func Create(Client *kclient.Client, name, size, componentName string) (*corev1.PersistentVolumeClaim, error) {

	labels := map[string]string{
		"component":    componentName,
		"storage-name": name,
	}

	quantity, err := resource.ParseQuantity(size)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse size: %v", size)
	}

	randomChars := util.GenerateRandomString(4)
	namespaceKubernetesObject, err := util.NamespaceOpenShiftObject(name, componentName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create namespaced name")
	}
	namespaceKubernetesObject = fmt.Sprintf("%v-%v", namespaceKubernetesObject, randomChars)

	objectMeta := kclient.CreateObjectMeta(namespaceKubernetesObject, Client.Namespace, labels, nil)
	pvcSpec := kclient.GeneratePVCSpec(quantity)

	// Create PVC
	glog.V(3).Infof("Creating a PVC with name %v\n", namespaceKubernetesObject)
	pvc, err := Client.CreatePVC(objectMeta, *pvcSpec)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create PVC")
	}
	return pvc, nil
}
