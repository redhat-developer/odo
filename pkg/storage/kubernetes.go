package storage

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/devfile/library/pkg/devfile/generator"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/occlient"
	storagelabels "github.com/openshift/odo/pkg/storage/labels"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// kubernetesClient contains information required for devfile based Storage operations
type kubernetesClient struct {
	generic
	client occlient.Client
}

// Create creates a pvc from the given Storage
func (k kubernetesClient) Create(storage Storage) error {

	pvcName, err := generatePVCName(storage.Name, k.componentName, k.appName)
	if err != nil {
		return err
	}

	labels := storagelabels.GetLabels(storage.Name, k.componentName, k.appName, true)

	labels["component"] = k.componentName
	labels[storagelabels.DevfileStorageLabel] = storage.Name

	if strings.Contains(storage.Name, OdoSourceVolume) {
		// Add label for source pvc
		labels[storagelabels.SourcePVCLabel] = storage.Name
	}

	objectMeta := generator.GetObjectMeta(pvcName, k.client.GetKubeClient().Namespace, labels, nil)

	quantity, err := resource.ParseQuantity(storage.Spec.Size)
	if err != nil {
		return errors.Wrapf(err, "unable to parse size: %v", storage.Spec.Size)
	}

	pvcParams := generator.PVCParams{
		ObjectMeta: objectMeta,
		Quantity:   quantity,
	}
	pvc := generator.GetPVC(pvcParams)

	// Create PVC
	klog.V(2).Infof("Creating a PVC with name %v and labels %v", pvcName, labels)
	_, err = k.client.GetKubeClient().CreatePVC(*pvc)
	if err != nil {
		return errors.Wrap(err, "unable to create PVC")
	}
	return nil
}

// Delete deletes the pvc belonging to the given Storage
func (k kubernetesClient) Delete(name string) error {
	pvcName, err := getPVCNameFromStorageName(&k.client, name)
	if err != nil {
		return err
	}

	// delete the associated PVC with the component
	err = k.client.GetKubeClient().DeletePVC(pvcName)
	if err != nil {
		return errors.Wrapf(err, "unable to delete PVC %v", pvcName)
	}

	return nil
}

// ListFromCluster lists pvc based Storage from the cluster
func (k kubernetesClient) ListFromCluster() (StorageList, error) {
	pod, err := k.client.GetKubeClient().GetOnePod(k.localConfig.GetName(), k.localConfig.GetApplication())
	if err != nil {
		if _, ok := err.(*kclient.PodNotFoundError); ok {
			return StorageList{}, nil
		}
		return StorageList{}, err
	}

	initContainerVolumeMounts := make(map[string]bool)
	for _, container := range pod.Spec.InitContainers {
		for _, volumeMount := range container.VolumeMounts {
			initContainerVolumeMounts[volumeMount.Name] = true
		}
	}

	var storage []Storage
	var volumeMounts []Storage
	for _, container := range pod.Spec.Containers {
		for _, volumeMount := range container.VolumeMounts {

			// avoid the volume mounts from the init containers
			// and the source volume volume mount
			_, ok := initContainerVolumeMounts[volumeMount.Name]
			if ok || volumeMount.Name == OdoSourceVolume {
				continue
			}

			volumeMounts = append(volumeMounts, Storage{
				ObjectMeta: metav1.ObjectMeta{Name: volumeMount.Name},
				Spec: StorageSpec{
					Path:          volumeMount.MountPath,
					ContainerName: container.Name,
				},
			})

		}
	}

	if len(volumeMounts) <= 0 {
		return StorageList{}, nil
	}

	selector := fmt.Sprintf("%v=%s,%s!=odo-projects", "component", k.localConfig.GetName(), storagelabels.SourcePVCLabel)

	pvcs, err := k.client.GetKubeClient().ListPVCs(selector)
	if err != nil {
		return StorageList{}, errors.Wrapf(err, "unable to get PVC using selector %v", storagelabels.StorageLabel)
	}

	// to track volume mounts used by a PVC
	validVolumeMounts := make(map[string]bool)

	for _, pvc := range pvcs {
		found := false
		for _, volumeMount := range volumeMounts {
			if volumeMount.Name == pvc.Name+"-vol" {
				// this volume mount is used by a PVC
				validVolumeMounts[volumeMount.Name] = true

				found = true
				size := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
				storage = append(storage, GetMachineFormatWithContainer(pvc.Labels[storagelabels.DevfileStorageLabel], size.String(), volumeMount.Spec.Path, volumeMount.Spec.ContainerName))
			}
		}
		if !found {
			return StorageList{}, fmt.Errorf("mount path for pvc %s not found", pvc.Name)
		}
	}

	// to track volumes created by Service Binding Operator
	for _, volume := range pod.Spec.Volumes {
		if volume.Secret != nil {
			validVolumeMounts[volume.Name] = true
		}
	}

	for _, volumeMount := range volumeMounts {
		if _, ok := validVolumeMounts[volumeMount.Name]; !ok {
			return StorageList{}, fmt.Errorf("pvc not found for mount path %s", volumeMount.Name)
		}
	}

	return StorageList{Items: storage}, nil
}

// List lists pvc based Storage and local Storage with respective states
func (k kubernetesClient) List() (StorageList, error) {
	localConfigStorage, err := k.localConfig.ListStorage()
	if err != nil {
		return StorageList{}, err
	}

	localStorage := ConvertListLocalToMachine(localConfigStorage)
	var clusterStorage StorageList
	if k.client.GetKubeClient() != nil {
		clusterStorage, err = k.ListFromCluster()
		if err != nil {
			return StorageList{}, err
		}
	}

	var storageList []Storage

	// find the local storage which are in a pushed and not pushed state
	for _, localStore := range localStorage.Items {
		found := false
		for _, clusterStore := range clusterStorage.Items {
			if reflect.DeepEqual(localStore, clusterStore) {
				found = true
			}
		}
		if found {
			localStore.Status = StateTypePushed
		} else {
			localStore.Status = StateTypeNotPushed
		}
		storageList = append(storageList, localStore)
	}

	// find the cluster storage which have been deleted locally
	for _, clusterStore := range clusterStorage.Items {
		found := false
		for _, localStore := range localStorage.Items {
			if reflect.DeepEqual(localStore, clusterStore) {
				found = true
			}
		}
		if !found {
			clusterStore.Status = StateTypeLocallyDeleted
			storageList = append(storageList, clusterStore)
		}
	}
	return GetMachineReadableFormatForList(storageList), nil
}
