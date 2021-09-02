package storage

import (
	"fmt"

	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/occlient"
	storagelabels "github.com/openshift/odo/pkg/storage/labels"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

// s2iClient contains information required for s2i based Storage based operations
type s2iClient struct {
	generic
	client occlient.Client
}

// this method is currently not being used by s2i components
// it is here to satisfy the interface
func (s s2iClient) Create(storage Storage) error {
	return nil
}

// this method is currently not being used by s2i components
// it is here to satisfy the interface
func (s s2iClient) Delete(name string) error {
	return nil
}

// ListFromCluster lists pvc based Storage from the cluster for s2i components
func (s s2iClient) ListFromCluster() (StorageList, error) {
	componentLabels := componentlabels.GetLabels(s.localConfig.GetName(), s.localConfig.GetApplication(), false)
	componentSelector := util.ConvertLabelsToSelector(componentLabels)

	dc, err := s.client.GetDeploymentConfigFromSelector(componentSelector)
	if err != nil {
		return StorageList{}, errors.Wrapf(err, "unable to get Deployment Config associated with component %v", s.localConfig.GetName())
	}

	pvcs, err := s.client.GetKubeClient().ListPVCs(storagelabels.StorageLabel)
	if err != nil {
		return StorageList{}, errors.Wrapf(err, "unable to get PVC using selector %v", storagelabels.StorageLabel)
	}

	pvcMap := make(map[string]*corev1.PersistentVolumeClaim)
	// store in map for faster searching
	for _, pvc := range pvcs {
		readPVC := pvc
		pvcMap[pvc.Name] = &readPVC
	}

	// store the storage in a map for faster searching with the key instead of list
	mountedStorageMap := make(map[string]string)
	volumeMounts := s.client.GetVolumeMountsFromDC(dc)
	for _, volumeMount := range volumeMounts {

		// We should ignore emptyDir (related to supervisord implementation)
		if s.client.IsVolumeAnEmptyDir(volumeMount.Name, dc) {
			continue
		}

		// We should ignore ConfigMap (while PR2142 and PR2601 are not fixed)
		if s.client.IsVolumeAnConfigMap(volumeMount.Name, dc) {
			continue
		}

		pvcName := s.client.GetPVCNameFromVolumeMountName(volumeMount.Name, dc)
		if pvcName == "" {
			return StorageList{}, fmt.Errorf("no PVC associated with Volume Mount %v", volumeMount.Name)
		}

		pvc, ok := pvcMap[pvcName]
		if !ok {
			// since the pvc doesn't exist, it might be a supervisorD volume
			// if true, continue
			if s.client.IsAppSupervisorDVolume(volumeMount.Name, dc.Name) {
				continue
			}
			return StorageList{}, fmt.Errorf("unable to get PVC %v", pvcName)
		}

		storageName := getStorageFromPVC(pvc)
		mountedStorageMap[storageName] = volumeMount.MountPath
	}

	var storage []Storage
	for i := range pvcs {
		pvc := pvcs[i]
		pvcComponentName, ok := pvc.Labels[componentlabels.ComponentLabel]
		pvcAppName, okApp := pvc.Labels[applabels.ApplicationLabel]
		// first check if component label does not exists indicating that the storage is not mounted in any component
		// if the component label exists, then check if the component is the current active component
		// also check if the app label exists and is equal to the current application
		if (!ok || pvcComponentName == s.localConfig.GetName()) && (okApp && pvcAppName == s.localConfig.GetApplication()) {
			if pvc.Name == "" {
				return StorageList{}, fmt.Errorf("no PVC associated")
			}
			storageName := getStorageFromPVC(&pvc)
			storageSize := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
			storageMachineReadable := NewStorage(getStorageFromPVC(&pvc),
				storageSize.String(),
				mountedStorageMap[storageName],
			)
			storage = append(storage, storageMachineReadable)
		}
	}
	storageList := NewStorageList(storage)
	return storageList, nil
}

// List lists pvc based Storage and local Storage with respective states for s2i components
func (s s2iClient) List() (StorageList, error) {

	storageConfig, err := s.localConfig.ListStorage()
	if err != nil {
		return StorageList{}, err
	}

	storageListConfig := ConvertListLocalToMachine(storageConfig)

	storageCluster, err := s.ListFromCluster()
	if err != nil {
		klog.V(4).Infof("Storage list from cluster error: %v", err)
	}

	var storageList []Storage

	// Iterate over local storage list, to add State PUSHED/NOT PUSHED
	for _, storeLocal := range storageListConfig.Items {
		storeLocal.Status = StateTypeNotPushed
		if isPushed(storeLocal.Name, storageCluster) {
			storeLocal.Status = StateTypePushed
		}
		storageList = append(storageList, storeLocal)
	}

	// Iterate over cluster storage list, to add State Locally Deleted
	for _, storeCluster := range storageCluster.Items {
		if isLocallyDeleted(storeCluster.Name, storageListConfig) {
			storeCluster.Status = StateTypeLocallyDeleted
			storageList = append(storageList, storeCluster)
		}
	}

	return NewStorageList(storageList), nil
}
