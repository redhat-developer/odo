package occlient

import (
	"fmt"
	appsv1 "github.com/openshift/api/apps/v1"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// CreatePVC creates a PVC resource in the cluster with the given name, size and
// labels
func (c *Client) CreatePVC(name string, size string, labels map[string]string, ownerReference ...metav1.OwnerReference) (*corev1.PersistentVolumeClaim, error) {
	quantity, err := resource.ParseQuantity(size)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse size: %v", size)
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: quantity,
				},
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
		},
	}

	for _, owRf := range ownerReference {
		pvc.SetOwnerReferences(append(pvc.GetOwnerReferences(), owRf))
	}

	createdPvc, err := c.kubeClient.CoreV1().PersistentVolumeClaims(c.Namespace).Create(pvc)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create PVC")
	}
	return createdPvc, nil
}

// AddPVCToDeploymentConfig adds the given PVC to the given Deployment Config
// at the given path
func (c *Client) AddPVCToDeploymentConfig(dc *appsv1.DeploymentConfig, pvc string, path string) error {
	volumeName := generateVolumeNameFromPVC(pvc)

	// Validating dc.Spec.Template is present before dereferencing
	if dc.Spec.Template == nil {
		return fmt.Errorf("TemplatePodSpec in %s DeploymentConfig is empty", dc.Name)
	}
	dc.Spec.Template.Spec.Volumes = append(dc.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvc,
			},
		},
	})

	// Validating dc.Spec.Template.Spec.Containers[] is present before dereferencing
	if len(dc.Spec.Template.Spec.Containers) == 0 {
		return fmt.Errorf("DeploymentConfig %s doesn't have any Containers defined", dc.Name)
	}
	dc.Spec.Template.Spec.Containers[0].VolumeMounts = append(dc.Spec.Template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
		Name:      volumeName,
		MountPath: path,
	},
	)
	return nil
}

// UpdatePVCLabels updates the given PVC with the given labels
func (c *Client) UpdatePVCLabels(pvc *corev1.PersistentVolumeClaim, labels map[string]string) error {
	pvc.Labels = labels
	_, err := c.kubeClient.CoreV1().PersistentVolumeClaims(c.Namespace).Update(pvc)
	if err != nil {
		return errors.Wrap(err, "unable to remove storage label from PVC")
	}
	return nil
}

// DeletePVC deletes the given PVC by name
func (c *Client) DeletePVC(name string) error {
	return c.kubeClient.CoreV1().PersistentVolumeClaims(c.Namespace).Delete(name, nil)
}

// IsAppSupervisorDVolume checks if the volume is a supervisorD volume
func (c *Client) IsAppSupervisorDVolume(volumeName, dcName string) bool {
	return volumeName == getAppRootVolumeName(dcName)
}

// getVolumeNamesFromPVC returns the name of the volume associated with the given
// PVC in the given Deployment Config
func (c *Client) getVolumeNamesFromPVC(pvc string, dc *appsv1.DeploymentConfig) []string {
	var volumes []string
	for _, volume := range dc.Spec.Template.Spec.Volumes {

		// If PVC does not exist, we skip (as this is either EmptyDir or "shared-data" from SupervisorD
		if volume.PersistentVolumeClaim == nil {
			klog.V(4).Infof("Volume has no PVC, skipping %s", volume.Name)
			continue
		}

		// If we find the PVC, add to volumes to be returned
		if volume.PersistentVolumeClaim.ClaimName == pvc {
			volumes = append(volumes, volume.Name)
		}

	}
	return volumes
}

// removeVolumeFromDC removes the volume from the given Deployment Config and
// returns true. If the given volume is not found, it returns false.
func removeVolumeFromDC(vol string, dc *appsv1.DeploymentConfig) bool {
	found := false
	for i, volume := range dc.Spec.Template.Spec.Volumes {
		if volume.Name == vol {
			found = true
			dc.Spec.Template.Spec.Volumes = append(dc.Spec.Template.Spec.Volumes[:i], dc.Spec.Template.Spec.Volumes[i+1:]...)
		}
	}
	return found
}

// removeVolumeMountsFromDC removes the volumeMounts from all the given containers
// in the given Deployment Config and return true. If any of the volumeMount with the name
// is not found, it returns false.
func removeVolumeMountsFromDC(vm string, dc *appsv1.DeploymentConfig) bool {
	found := false
	for i, container := range dc.Spec.Template.Spec.Containers {
		for j, volumeMount := range container.VolumeMounts {
			if volumeMount.Name == vm {
				found = true
				dc.Spec.Template.Spec.Containers[i].VolumeMounts = append(dc.Spec.Template.Spec.Containers[i].VolumeMounts[:j], dc.Spec.Template.Spec.Containers[i].VolumeMounts[j+1:]...)
			}
		}
	}
	return found
}

// generateVolumeNameFromPVC generates a random volume name based on the name
// of the given PVC
func generateVolumeNameFromPVC(pvc string) string {
	return fmt.Sprintf("%v-%v-volume", pvc, util.GenerateRandomString(nameLength))
}

// addOrRemoveVolumeAndVolumeMount mounts or unmounts PVCs from the given deploymentConfig
func addOrRemoveVolumeAndVolumeMount(client *Client, dc *appsv1.DeploymentConfig, storageToMount map[string]*corev1.PersistentVolumeClaim, storageUnMount map[string]string) error {

	if len(dc.Spec.Template.Spec.Containers) == 0 || len(dc.Spec.Template.Spec.Containers) > 1 {
		return fmt.Errorf("more than one container found in dc")
	}

	// find the volume mount to be unmounted from the dc
	for i, volumeMount := range dc.Spec.Template.Spec.Containers[0].VolumeMounts {
		if _, ok := storageUnMount[volumeMount.MountPath]; ok {
			dc.Spec.Template.Spec.Containers[0].VolumeMounts = append(dc.Spec.Template.Spec.Containers[0].VolumeMounts[:i], dc.Spec.Template.Spec.Containers[0].VolumeMounts[i+1:]...)

			// now find the volume to be deleted from the dc
			for j, volume := range dc.Spec.Template.Spec.Volumes {
				if volume.Name == volumeMount.Name {
					dc.Spec.Template.Spec.Volumes = append(dc.Spec.Template.Spec.Volumes[:j], dc.Spec.Template.Spec.Volumes[j+1:]...)
				}
			}
		}
	}

	for path, pvc := range storageToMount {
		err := client.AddPVCToDeploymentConfig(dc, pvc.Name, path)
		if err != nil {
			return errors.Wrap(err, "unable to add pvc to deployment config")
		}
	}
	return nil
}

// updateStorageOwnerReference updates the given storage with the given owner references
func updateStorageOwnerReference(client *Client, pvc *corev1.PersistentVolumeClaim, ownerReference ...metav1.OwnerReference) error {
	if len(ownerReference) <= 0 {
		return errors.New("owner references are empty")
	}
	// get the latest version of the PVC to avoid conflict errors
	latestPVC, err := client.kubeClient.CoreV1().PersistentVolumeClaims(client.Namespace).Get(pvc.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	for _, owRf := range ownerReference {
		latestPVC.SetOwnerReferences(append(pvc.GetOwnerReferences(), owRf))
	}
	_, err = client.kubeClient.CoreV1().PersistentVolumeClaims(client.Namespace).Update(latestPVC)
	if err != nil {
		return err
	}
	return nil
}
