package storage

import (
	"fmt"

	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"

	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/occlient"
	storagelabels "github.com/openshift/odo/pkg/storage/labels"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const (
	// OdoSourceVolume is the constant containing the name of the emptyDir volume containing the project source
	OdoSourceVolume = "odo-projects"

	// OdoSourceVolumeSize specifies size for odo source volume.
	OdoSourceVolumeSize = "2Gi"

	apiVersion = "odo.dev/v1alpha1"
)

// Get returns Storage defination for given Storage name
func (storages StorageList) Get(storageName string) Storage {
	for _, storage := range storages.Items {
		if storage.Name == storageName {
			return storage
		}
	}
	return Storage{}

}

// Create adds storage to given component of given application
func Create(client *occlient.Client, name string, size string, componentName string, applicationName string) (*corev1.PersistentVolumeClaim, error) {

	// Namespace the component
	// We will use name+applicationName instead of componentName+applicationName until:
	// https://github.com/openshift/odo/issues/504 is resolved.
	namespacedOpenShiftObject, err := util.NamespaceOpenShiftObject(name, applicationName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create namespaced name")
	}

	labels := storagelabels.GetLabels(name, componentName, applicationName, true)

	klog.V(4).Infof("Got labels for PVC: %v", labels)

	// Create PVC
	pvc, err := client.CreatePVC(generatePVCNameFromStorageName(namespacedOpenShiftObject), size, labels)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create PVC")
	}
	return pvc, nil
}

// Unmount unmounts the given storage from the given component
// updateLabels is a flag to whether update Label or not, so updation of label
// is not required in delete call but required in unmount call
// this is introduced as causing unnecessary delays
func Unmount(client *occlient.Client, storageName string, componentName string, applicationName string, updateLabels bool) error {
	// Get DeploymentConfig for the given component
	componentLabels := componentlabels.GetLabels(componentName, applicationName, false)
	componentSelector := util.ConvertLabelsToSelector(componentLabels)
	dc, err := client.GetDeploymentConfigFromSelector(componentSelector)
	if err != nil {
		return errors.Wrapf(err, "unable to get Deployment Config for component: %v in application: %v", componentName, applicationName)
	}

	pvcName, err := getPVCNameFromStorageName(client, storageName)
	if err != nil {
		return errors.Wrapf(err, "unable to get PVC for storage %v", storageName)
	}

	// Remove PVC from Deployment Config
	if err := client.RemoveVolumeFromDeploymentConfig(pvcName, dc.Name); err != nil {
		return errors.Wrapf(err, "unable to remove volume: %v from Deployment Config: %v", pvcName, dc.Name)
	}

	pvc, err := client.GetKubeClient().GetPVCFromName(pvcName)
	if err != nil {
		return errors.Wrapf(err, "unable to get PersistentVolumeClaim from name: %v", pvcName)
	}
	pvcLabels := applabels.GetLabels(applicationName, true)
	pvcLabels[storagelabels.StorageLabel] = storageName

	if updateLabels {
		if err := client.GetKubeClient().UpdatePVCLabels(pvc, pvcLabels); err != nil {
			return errors.Wrapf(err, "unable to remove storage label from : %v", pvc.Name)
		}
	}
	return nil
}

// Delete removes storage from the given application.
// Delete returns the component name, if it is mounted to a component, or "" and the error, if any
func Delete(client *occlient.Client, name string) error {
	pvcName, err := getPVCNameFromStorageName(client, name)
	if err != nil {
		return errors.Wrapf(err, "unable to get PVC for storage %v", name)
	}

	// delete the associated PVC with the component
	err = client.GetKubeClient().DeletePVC(pvcName)
	if err != nil {
		return errors.Wrapf(err, "unable to delete PVC %v", pvcName)
	}

	return nil
}

// List lists all the mounted storage associated with the given component of the given
// application and the unmounted storage in the given application
func List(client *occlient.Client, componentName string, applicationName string) (StorageList, error) {
	componentLabels := componentlabels.GetLabels(componentName, applicationName, false)
	componentSelector := util.ConvertLabelsToSelector(componentLabels)

	dc, err := client.GetDeploymentConfigFromSelector(componentSelector)
	if err != nil {
		return StorageList{}, errors.Wrapf(err, "unable to get Deployment Config associated with component %v", componentName)
	}

	pvcs, err := client.GetKubeClient().ListPVCs(storagelabels.StorageLabel)
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
	volumeMounts := client.GetVolumeMountsFromDC(dc)
	for _, volumeMount := range volumeMounts {

		// We should ignore emptyDir (related to supervisord implementation)
		if client.IsVolumeAnEmptyDir(volumeMount.Name, dc) {
			continue
		}

		// We should ignore ConfigMap (while PR2142 and PR2601 are not fixed)
		if client.IsVolumeAnConfigMap(volumeMount.Name, dc) {
			continue
		}

		pvcName := client.GetPVCNameFromVolumeMountName(volumeMount.Name, dc)
		if pvcName == "" {
			return StorageList{}, fmt.Errorf("no PVC associated with Volume Mount %v", volumeMount.Name)
		}

		pvc, ok := pvcMap[pvcName]
		if !ok {
			// since the pvc doesn't exist, it might be a supervisorD volume
			// if true, continue
			if client.IsAppSupervisorDVolume(volumeMount.Name, dc.Name) {
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
		if (!ok || pvcComponentName == componentName) && (okApp && pvcAppName == applicationName) {
			if pvc.Name == "" {
				return StorageList{}, fmt.Errorf("no PVC associated")
			}
			storageName := getStorageFromPVC(&pvc)
			storageSize := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
			storageMachineReadable := GetMachineReadableFormat(getStorageFromPVC(&pvc),
				storageSize.String(),
				mountedStorageMap[storageName],
			)
			storage = append(storage, storageMachineReadable)
		}
	}
	storageList := GetMachineReadableFormatForList(storage)
	return storageList, nil
}

// ListMounted lists all the mounted storage associated with the given component and application
func ListMounted(client *occlient.Client, componentName string, applicationName string) (StorageList, error) {
	storageList, err := List(client, componentName, applicationName)
	if err != nil {
		return StorageList{}, errors.Wrapf(err, "unable to get storage of component %v", componentName)
	}
	var storageListMounted []Storage
	for _, storage := range storageList.Items {
		if storage.Spec.Path != "" {
			storageListMounted = append(storageListMounted, storage)
		}
	}
	return GetMachineReadableFormatForList(storageListMounted), nil
}

// ListUnmounted lists all the unmounted storage associated with the given application
func ListUnmounted(client *occlient.Client, applicationName string) (StorageList, error) {
	pvcs, err := client.GetKubeClient().ListPVCs(storagelabels.StorageLabel)
	if err != nil {
		return StorageList{}, errors.Wrapf(err, "unable to get PVC using selector %v", storagelabels.StorageLabel)
	}
	var storage []Storage
	for i := range pvcs {
		pvc := pvcs[i]
		_, ok := pvc.Labels[componentlabels.ComponentLabel]
		pvcAppName, okApp := pvc.Labels[applabels.ApplicationLabel]
		// first check if component label does not exists indicating that the storage is not mounted in any component
		// also check if the app label exists and is equal to the current application
		if !ok && (okApp && pvcAppName == applicationName) {
			if pvc.Name == "" {
				return StorageList{}, fmt.Errorf("no PVC associated")
			}
			storageSize := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
			storageMachineReadable := GetMachineReadableFormat(getStorageFromPVC(&pvc),
				storageSize.String(),
				"",
			)
			storage = append(storage, storageMachineReadable)
		}
	}
	storageList := GetMachineReadableFormatForList(storage)
	return storageList, nil
}

// Exists checks if the given storage exists in the given application
func Exists(client *occlient.Client, storageName string, applicationName string) (bool, error) {
	var labels = make(map[string]string)
	labels[applabels.ApplicationLabel] = applicationName
	labels[storagelabels.StorageLabel] = storageName
	selector := util.ConvertLabelsToSelector(labels)
	pvcs, err := client.GetKubeClient().ListPVCs(selector)
	if err != nil {
		return false, errors.Wrapf(err, "unable to list storage for application %v", applicationName)
	}

	if len(pvcs) <= 0 {
		return false, nil
	}
	return true, nil
}

// generatePVCNameFromStorageName generates a PVC name from the given storage
// name
func generatePVCNameFromStorageName(storage string) string {
	return fmt.Sprintf("%v-pvc", storage)
}

// getStorageFromPVC returns the storage associated with the given PVC
func getStorageFromPVC(pvc *corev1.PersistentVolumeClaim) string {
	if _, ok := pvc.Labels[storagelabels.StorageLabel]; !ok {
		return ""
	}
	return pvc.Labels[storagelabels.StorageLabel]
}

// getPVCNameFromStorageName returns the PVC associated with the given storage
func getPVCNameFromStorageName(client *occlient.Client, storageName string) (string, error) {
	var labels = make(map[string]string)
	labels[storagelabels.StorageLabel] = storageName

	selector := util.ConvertLabelsToSelector(labels)
	pvcs, err := client.GetKubeClient().ListPVCNames(selector)
	if err != nil {
		return "", errors.Wrapf(err, "unable to get PVC names for selector %v", selector)
	}
	numPVCs := len(pvcs)
	if numPVCs != 1 {
		return "", fmt.Errorf("expected exactly one PVC attached to storage %v, but got %v, %v", storageName, numPVCs, pvcs)
	}
	return pvcs[0], nil
}

// GetComponentNameFromStorageName returns the component name associated with the storageName, if any, or ""
func GetComponentNameFromStorageName(client *occlient.Client, storageName string) (string, error) {
	var labels = make(map[string]string)
	labels[storagelabels.StorageLabel] = storageName

	selector := util.ConvertLabelsToSelector(labels)
	pvcs, err := client.GetKubeClient().ListPVCs(selector)
	if err != nil {
		return "", errors.Wrap(err, "unable to list the pvcs")
	}
	if len(pvcs) > 1 {
		return "", errors.Wrap(err, "more than one pvc found for the storage label")
	}
	if len(pvcs) == 1 {
		pvc := pvcs[0]
		labels = pvc.GetLabels()
		return labels[componentlabels.ComponentLabel], nil
	}
	return "", nil
}

// IsMounted checks if the given storage is mounted to the given component
// IsMounted returns a bool indicating the storage is mounted to the component or not
func IsMounted(client *occlient.Client, storageName string, componentName string, applicationName string) (bool, error) {
	storageList, err := List(client, componentName, applicationName)
	if err != nil {
		return false, errors.Wrapf(err, "unable to list storage for component %v", componentName)
	}
	for _, storage := range storageList.Items {
		if storage.Name == storageName {
			if storage.Spec.Path != "" {
				return true, nil
			}
		}
	}
	return false, nil
}

// Mount mounts the given storage to the given component
func Mount(client *occlient.Client, path string, storageName string, componentName string, applicationName string) error {
	storageComponent, err := GetComponentNameFromStorageName(client, storageName)
	if err != nil {
		return errors.Wrap(err, "unable to get the component name associated with the storage")
	}
	if storageComponent != "" {
		return fmt.Errorf("the given storage is already mounted to the component %v", storageComponent)
	}

	namespacedOpenShiftObject, err := util.NamespaceOpenShiftObject(storageName, applicationName)
	if err != nil {
		return errors.Wrapf(err, "unable to create namespaced name")
	}

	pvc, err := client.GetKubeClient().GetPVCFromName(generatePVCNameFromStorageName(namespacedOpenShiftObject))
	if err != nil {
		return errors.Wrap(err, "unable to get the pvc from the storage name")
	}

	// Get DeploymentConfig for the given component
	componentLabels := componentlabels.GetLabels(componentName, applicationName, false)
	componentSelector := util.ConvertLabelsToSelector(componentLabels)
	dc, err := client.GetDeploymentConfigFromSelector(componentSelector)
	if err != nil {
		return errors.Wrapf(err, "unable to get Deployment Config for component: %v in application: %v", componentName, applicationName)
	}
	klog.V(4).Infof("Deployment Config: %v is associated with the component: %v", dc.Name, componentName)

	// Add PVC to DeploymentConfig
	if err := client.AddPVCToDeploymentConfig(dc, pvc.Name, path); err != nil {
		return errors.Wrap(err, "unable to add PVC to DeploymentConfig")
	}
	err = client.GetKubeClient().UpdatePVCLabels(pvc, storagelabels.GetLabels(storageName, componentName, applicationName, true))
	if err != nil {
		return errors.Wrap(err, "unable to update the pvc")
	}
	return nil
}

// Gets the storageName mounted to the given path in the given component and application
// GetStorageNameFromMountPath returns the name of the storage or the error
func GetStorageNameFromMountPath(client *occlient.Client, path string, componentName string, applicationName string) (string, error) {
	storageList, err := List(client, componentName, applicationName)
	if err != nil {
		return "", errors.Wrapf(err, "unable to list storage for component %v", componentName)
	}
	for _, storage := range storageList.Items {
		if storage.Spec.Path == path {
			return storage.Name, nil
		}
	}
	return "", nil
}

// S2iPush creates/deletes the required storage during `odo push`
// storageList are the storage mentioned in the config
// isComponentExists indicates if the component exists or not, if not, we don't run the list operation
// returns the storage for mounting and unMounting from the DC
// StorageToBeMounted describes the storage to be mounted
// StorageToBeMounted : storagePath is the key of the map, the generatedPVC is the value of the map
// StorageToBeUnMounted describes the storage to be unmounted
// StorageToBeUnMounted : path is the key of the map,storageName is the value of the map
func S2iPush(client *occlient.Client, storageList StorageList, componentName, applicationName string, isComponentExits bool) (map[string]*corev1.PersistentVolumeClaim, map[string]string, error) {
	// list all the storage in the cluster
	storageClusterList := StorageList{}
	var err error
	if isComponentExits {
		storageClusterList, err = ListMounted(client, componentName, applicationName)

	}
	if err != nil {
		return nil, nil, err
	}
	storageClusterNames := make(map[string]Storage)
	for _, storage := range storageClusterList.Items {
		storageClusterNames[storage.Name] = storage
	}

	// list all the storage in the config
	storageConfigNames := make(map[string]Storage)
	for _, storage := range storageList.Items {
		storageConfigNames[storage.Name] = storage
	}

	storageToBeMounted := make(map[string]*corev1.PersistentVolumeClaim)
	storageToBeUnMounted := make(map[string]string)

	// find storage to delete
	for _, storage := range storageClusterList.Items {
		val, ok := storageConfigNames[storage.Name]
		if !ok {
			// delete the pvc
			err = Delete(client, storage.Name)
			if err != nil {
				return nil, nil, err
			}
			log.Successf("Deleted storage %v from %v", storage.Name, componentName)
			storageToBeUnMounted[storage.Spec.Path] = storage.Name
			continue
		} else if storage.Name == val.Name {
			if val.Spec.Size != storage.Spec.Size || val.Spec.Path != storage.Spec.Path {
				return nil, nil, errors.Errorf("config mismatch for storage with the same name %s", storage.Name)
			}
		}
	}

	// find storage to create
	for _, storage := range storageList.Items {
		_, ok := storageClusterNames[storage.Name]
		if !ok {
			createdPVC, err := Create(client, storage.Name, storage.Spec.Size, componentName, applicationName)
			if err != nil {
				return nil, nil, err
			}
			log.Successf("Added storage %v to %v", storage.Name, componentName)
			storageToBeMounted[storage.Spec.Path] = createdPVC
		}
	}

	return storageToBeMounted, storageToBeUnMounted, err
}

// GetMachineReadableFormatForList gives machine readable StorageList definition
func GetMachineReadableFormatForList(storage []Storage) StorageList {
	return StorageList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: apiVersion,
		},
		ListMeta: metav1.ListMeta{},
		Items:    storage,
	}
}

// GetMachineReadableFormat gives machine readable Storage definition
// storagePath indicates the path to which the storage is mounted to, "" if not mounted
func GetMachineReadableFormat(storageName, storageSize, storagePath string) Storage {
	return Storage{
		TypeMeta:   metav1.TypeMeta{Kind: "storage", APIVersion: apiVersion},
		ObjectMeta: metav1.ObjectMeta{Name: storageName},
		Spec: StorageSpec{
			Size: storageSize,
			Path: storagePath,
		},
	}
}

// GetMachineFormatWithContainer gives machine readable Storage definition
// storagePath indicates the path to which the storage is mounted to, "" if not mounted
func GetMachineFormatWithContainer(storageName, storageSize, storagePath string, container string) Storage {
	storage := Storage{
		TypeMeta:   metav1.TypeMeta{Kind: "storage", APIVersion: apiVersion},
		ObjectMeta: metav1.ObjectMeta{Name: storageName},
		Spec: StorageSpec{
			Size: storageSize,
			Path: storagePath,
		},
	}
	storage.Spec.ContainerName = container
	return storage
}

// generatePVCName generates a PVC name from the Devfile volume name, component name and app name
func generatePVCName(volName, componentName, appName string) (string, error) {

	pvcName, err := util.NamespaceKubernetesObject(volName, componentName)
	if err != nil {
		return "", errors.Wrapf(err, "unable to create namespaced name")
	}

	pvcName, err = util.NamespaceKubernetesObject(pvcName, appName)
	if err != nil {
		return "", errors.Wrapf(err, "unable to create namespaced name")
	}

	return pvcName, nil
}

func isLocallyDeleted(storageName string, storageLocal StorageList) bool {
	for _, storage := range storageLocal.Items {
		if storageName == storage.Name {
			return false
		}
	}

	return true
}

func isPushed(storageName string, storageCluster StorageList) bool {
	for _, storage := range storageCluster.Items {
		if storageName == storage.Name {
			return true
		}
	}

	return false
}

// It converts storage config list to StorageList type
func ConvertListLocalToMachine(storageListConfig []localConfigProvider.LocalStorage) StorageList {

	var storageListLocal []Storage

	for _, storeLocal := range storageListConfig {
		s := GetMachineReadableFormat(storeLocal.Name, storeLocal.Size, storeLocal.Path)
		s.Spec.ContainerName = storeLocal.Container
		storageListLocal = append(storageListLocal, s)
	}

	return GetMachineReadableFormatForList(storageListLocal)
}

// MachineReadableSuccessOutput outputs a success output that includes
// storage information
func MachineReadableSuccessOutput(storageName string, message string) {
	machineOutput := machineoutput.GenericSuccess{
		TypeMeta: metav1.TypeMeta{
			Kind:       "storage",
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: storageName,
		},
		Message: message,
	}

	machineoutput.OutputSuccess(machineOutput)
}

// DevfileListMounted lists the storage which are mounted on a container
func DevfileListMounted(kClient *kclient.Client, componentName, appName string) (StorageList, error) {
	pod, err := kClient.GetOnePod(componentName, appName)
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

	selector := fmt.Sprintf("component=%s,%s!=odo-projects", componentName, storagelabels.SourcePVCLabel)

	pvcs, err := kClient.ListPVCs(selector)
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

type ClientOptions struct {
	OCClient            occlient.Client
	LocalConfigProvider localConfigProvider.LocalConfigProvider
}

type Client interface {
	Create(Storage) error
	Delete(string) error
	ListFromCluster() (StorageList, error)
	List() (StorageList, error)
}

// NewClient gets the appropriate Storage client based on the parameters
func NewClient(options ClientOptions) Client {
	genericInfo := generic{
		appName:       options.LocalConfigProvider.GetApplication(),
		componentName: options.LocalConfigProvider.GetName(),
		localConfig:   options.LocalConfigProvider,
	}

	if _, ok := options.LocalConfigProvider.(*config.LocalConfigInfo); ok {
		return s2iClient{
			generic: genericInfo,
			client:  options.OCClient,
		}
	} else {
		return kubernetesClient{
			generic: genericInfo,
			client:  options.OCClient,
		}
	}
}

// generic contains information required for all the Storage clients
type generic struct {
	appName       string
	componentName string
	localConfig   localConfigProvider.LocalConfigProvider
}

// Push creates and deletes the required Storage
// it compares the local storage against the storage on the cluster
func Push(client Client, configProvider localConfigProvider.LocalConfigProvider) error {
	// list all the storage in the cluster
	storageClusterList := StorageList{}

	storageClusterList, err := client.ListFromCluster()
	if err != nil {
		return err
	}
	storageClusterNames := make(map[string]Storage)
	for _, storage := range storageClusterList.Items {
		storageClusterNames[storage.Name] = storage
	}

	// list all the storage in the config
	storageConfigNames := make(map[string]Storage)

	localStorage, err := configProvider.ListStorage()
	if err != nil {
		return err
	}
	for _, storage := range ConvertListLocalToMachine(localStorage).Items {
		storageConfigNames[storage.Name] = storage
	}

	// find storage to delete
	for storageName, storage := range storageClusterNames {
		val, ok := storageConfigNames[storageName]
		if !ok {
			// delete the pvc
			err = client.Delete(storage.Name)
			if err != nil {
				return err
			}
			log.Successf("Deleted storage %v from %v", storage.Name, configProvider.GetName())
			continue
		} else if storage.Name == val.Name {
			if val.Spec.Size != storage.Spec.Size {
				return errors.Errorf("config mismatch for storage with the same name %s", storage.Name)
			}
		}
	}

	// find storage to create
	for storageName, storage := range storageConfigNames {
		_, ok := storageClusterNames[storageName]
		if !ok {
			err := client.Create(storage)
			if err != nil {
				return err
			}
			log.Successf("Added storage %v to %v", storage.Name, configProvider.GetName())
		}
	}

	return err
}
