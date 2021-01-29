package storage

import (
	"fmt"
	"reflect"

	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/occlient"
	storagelabels "github.com/openshift/odo/pkg/storage/labels"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// kubernetesClient contains information required for devfile based Storage operations
type kubernetesClient struct {
	generic
	client occlient.Client
}

// ListFromCluster lists pvc based Storage from the cluster
func (k kubernetesClient) ListFromCluster() (StorageList, error) {
	pod, err := k.client.GetKubeClient().GetPodUsingComponentName(k.localConfig.GetName())
	if err != nil {
		if _, ok := err.(*kclient.PodNotFoundError); ok {
			return StorageList{}, nil
		}
		return StorageList{}, err
	}

	var storage []Storage
	var volumeMounts []Storage
	for _, container := range pod.Spec.Containers {
		for _, volumeMount := range container.VolumeMounts {

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

	selector := fmt.Sprintf("component=%s,%s!=odo-projects", k.localConfig.GetName(), storagelabels.SourcePVCLabel)

	pvcs, err := k.client.GetKubeClient().ListPVCs(selector)
	if err != nil {
		return StorageList{}, errors.Wrapf(err, "unable to get PVC using selector %v", storagelabels.StorageLabel)
	}

	for _, pvc := range pvcs {
		found := false
		for _, volumeMount := range volumeMounts {
			if volumeMount.Name == pvc.Name+"-vol" {
				found = true
				size := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
				storage = append(storage, GetMachineFormatWithContainer(pvc.Labels[storagelabels.DevfileStorageLabel], size.String(), volumeMount.Spec.Path, volumeMount.Spec.ContainerName))
			}
		}
		if !found {
			return StorageList{}, fmt.Errorf("mount path for pvc %s not found", pvc.Name)
		}
	}

	return StorageList{Items: storage}, nil
}

// List lists pvc based Storage and local Storage with respective states
func (k kubernetesClient) List() (StorageList, error) {
	localStorage := ConvertListLocalToMachine(k.localConfig.ListStorage())

	clusterStorage, err := k.ListFromCluster()
	if err != nil {
		return StorageList{}, err
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
