package storage

import (
	"fmt"
	"strings"

	"github.com/devfile/library/v2/pkg/devfile/generator"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
)

// kubernetesClient contains information required for devfile based Storage operations
type kubernetesClient struct {
	generic
	client kclient.ClientInterface

	// if we don't have access to the local config
	// we can use the deployment to call List() and
	// directly list storage from the cluster without the local config
	deployment *v1.Deployment
}

var _ Client = (*kubernetesClient)(nil)

// Create creates a pvc from the given Storage
func (k kubernetesClient) Create(storage Storage) error {

	if k.componentName == "" || k.appName == "" {
		return fmt.Errorf("the component name and the app name should be provided")
	}

	pvcName, err := generatePVCName(storage.Name, k.componentName, k.appName)
	if err != nil {
		return err
	}

	labels := odolabels.GetLabels(k.componentName, k.appName, k.runtime, odolabels.ComponentDevMode, false)
	odolabels.AddStorageInfo(labels, storage.Name, strings.Contains(storage.Name, OdoSourceVolume))

	objectMeta := generator.GetObjectMeta(pvcName, k.client.GetCurrentNamespace(), labels, nil)

	quantity, err := resource.ParseQuantity(storage.Spec.Size)
	if err != nil {
		return fmt.Errorf("unable to parse size: %v: %w", storage.Spec.Size, err)
	}

	pvcParams := generator.PVCParams{
		ObjectMeta: objectMeta,
		Quantity:   quantity,
	}
	pvc := generator.GetPVC(pvcParams)

	// Create PVC
	klog.V(2).Infof("Creating a PVC with name %v and labels %v", pvcName, labels)
	_, err = k.client.CreatePVC(*pvc)
	if err != nil {
		return fmt.Errorf("unable to create PVC: %w", err)
	}
	return nil
}

// Delete deletes the pvc belonging to the given Storage
func (k kubernetesClient) Delete(name string) error {
	pvcName, err := getPVCNameFromStorageName(k.client, name)
	if err != nil {
		return err
	}

	// delete the associated PVC with the component
	err = k.client.DeletePVC(pvcName)
	if err != nil {
		return fmt.Errorf("unable to delete PVC %v: %w", pvcName, err)
	}

	return nil
}

// List lists pvc based Storage from the cluster
func (k kubernetesClient) List() (StorageList, error) {
	if k.deployment == nil {
		var err error
		selector := odolabels.GetSelector(k.componentName, k.appName, odolabels.ComponentAnyMode, true)
		k.deployment, err = k.client.GetOneDeploymentFromSelector(selector)
		if err != nil {
			if _, ok := err.(*kclient.DeploymentNotFoundError); ok {
				return StorageList{}, nil
			}
			return StorageList{}, err
		}
	}

	initContainerVolumeMounts := make(map[string]bool)
	for _, container := range k.deployment.Spec.Template.Spec.InitContainers {
		for _, volumeMount := range container.VolumeMounts {
			initContainerVolumeMounts[volumeMount.Name] = true
		}
	}

	containerVolumeMounts := make(map[string]bool)
	for _, container := range k.deployment.Spec.Template.Spec.Containers {
		for _, volumeMount := range container.VolumeMounts {
			containerVolumeMounts[volumeMount.Name] = true
		}
	}

	var storage []Storage
	var volumeMounts []Storage
	for _, container := range k.deployment.Spec.Template.Spec.Containers {
		for _, volumeMount := range container.VolumeMounts {

			// avoid the volume mounts only from the init containers
			// and the source volume mount
			_, initOK := initContainerVolumeMounts[volumeMount.Name]
			_, ok := containerVolumeMounts[volumeMount.Name]
			if (!ok && initOK) || volumeMount.Name == OdoSourceVolume || volumeMount.Name == SharedDataVolumeName {
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

	selector := odolabels.SelectorBuilder().WithComponent(k.componentName).WithoutSourcePVC(OdoSourceVolume).Selector()
	pvcs, err := k.client.ListPVCs(selector)
	if err != nil {
		return StorageList{}, fmt.Errorf("unable to get PVC using selector %q: %w", selector, err)
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
				storage = append(storage, NewStorageWithContainer(odolabels.GetDevfileStorageName(pvc.Labels), size.String(), volumeMount.Spec.Path, volumeMount.Spec.ContainerName, nil))
			}
		}
		if !found {
			return StorageList{}, fmt.Errorf("mount path for pvc %s not found", pvc.Name)
		}
	}

	// to track non PVC volumes
	for _, volume := range k.deployment.Spec.Template.Spec.Volumes {
		if volume.PersistentVolumeClaim == nil {
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
