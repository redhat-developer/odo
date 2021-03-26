package occlient

import (
	"context"
	"fmt"

	"github.com/devfile/library/pkg/devfile/generator"
	appsv1 "github.com/openshift/api/apps/v1"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
)

// CreatePVC creates a PVC resource in the cluster with the given name, size and
// labels
func (c *Client) CreatePVC(name string, size string, labels map[string]string, ownerReference ...metav1.OwnerReference) (*corev1.PersistentVolumeClaim, error) {
	quantity, err := resource.ParseQuantity(size)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse size: %v", size)
	}

	pvcParams := generator.PVCParams{
		ObjectMeta: generator.GetObjectMeta(name, c.Namespace, labels, nil),
		Quantity:   quantity,
	}

	pvc := generator.GetPVC(pvcParams)

	for _, owRf := range ownerReference {
		pvc.SetOwnerReferences(append(pvc.GetOwnerReferences(), owRf))
	}

	return c.kubeClient.CreatePVC(*pvc)
}

// GetPVCNameFromVolumeMountName returns the PVC associated with the given volume
// An empty string is returned if the volume is not found
func (c *Client) GetPVCNameFromVolumeMountName(volumeMountName string, dc *appsv1.DeploymentConfig) string {
	for _, volume := range dc.Spec.Template.Spec.Volumes {
		if volume.Name == volumeMountName {
			if volume.PersistentVolumeClaim != nil {
				return volume.PersistentVolumeClaim.ClaimName
			}
		}
	}
	return ""
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

// IsAppSupervisorDVolume checks if the volume is a supervisorD volume
func (c *Client) IsAppSupervisorDVolume(volumeName, dcName string) bool {
	return volumeName == getAppRootVolumeName(dcName)
}

// IsVolumeAnEmptyDir returns true if the volume is an EmptyDir, false if not
func (c *Client) IsVolumeAnEmptyDir(volumeMountName string, dc *appsv1.DeploymentConfig) bool {
	for _, volume := range dc.Spec.Template.Spec.Volumes {
		if volume.Name == volumeMountName {
			if volume.EmptyDir != nil {
				return true
			}
		}
	}
	return false
}

// IsVolumeAnConfigMap returns true if the volume is an ConfigMap, false if not
func (c *Client) IsVolumeAnConfigMap(volumeMountName string, dc *appsv1.DeploymentConfig) bool {
	for _, volume := range dc.Spec.Template.Spec.Volumes {
		if volume.Name == volumeMountName {
			if volume.ConfigMap != nil {
				return true
			}
		}
	}
	return false
}

// RemoveVolumeFromDeploymentConfig removes the volume associated with the
// given PVC from the Deployment Config. Both, the volume entry and the
// volume mount entry in the containers, are deleted.
func (c *Client) RemoveVolumeFromDeploymentConfig(pvc string, dcName string) error {

	retryErr := retry.RetryOnConflict(retry.DefaultBackoff, func() error {

		dc, err := c.GetDeploymentConfigFromName(dcName)
		if err != nil {
			return errors.Wrapf(err, "unable to get Deployment Config: %v", dcName)
		}

		volumeNames := getVolumeNamesFromPVC(pvc, dc)
		numVolumes := len(volumeNames)
		if numVolumes == 0 {
			return fmt.Errorf("no volume found for PVC %v in DC %v, expected one", pvc, dc.Name)
		} else if numVolumes > 1 {
			return fmt.Errorf("found more than one volume for PVC %v in DC %v, expected one", pvc, dc.Name)
		}
		volumeName := volumeNames[0]

		// Remove volume if volume exists in Deployment Config
		err = removeVolumeFromDC(volumeName, dc)
		if err != nil {
			return err
		}
		klog.V(3).Infof("Found volume: %v in Deployment Config: %v", volumeName, dc.Name)

		// Remove at max 2 volume mounts if volume mounts exists
		err = removeVolumeMountsFromDC(volumeName, dc)
		if err != nil {
			return err
		}

		_, updateErr := c.appsClient.DeploymentConfigs(c.Namespace).Update(context.TODO(), dc, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		return errors.Wrapf(retryErr, "updating Deployment Config %v failed", dcName)
	}
	return nil
}

// getVolumeNamesFromPVC returns the name of the volume associated with the given
// PVC in the given Deployment Config
func getVolumeNamesFromPVC(pvc string, dc *appsv1.DeploymentConfig) []string {
	var volumes []string
	for _, volume := range dc.Spec.Template.Spec.Volumes {

		// If PVC does not exist, we skip (as this is either EmptyDir or "shared-data" from SupervisorD
		if volume.PersistentVolumeClaim == nil {
			klog.V(3).Infof("Volume has no PVC, skipping %s", volume.Name)
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
func removeVolumeFromDC(vol string, dc *appsv1.DeploymentConfig) error {

	// Error out immediately if there are zero volumes to begin with
	if len(dc.Spec.Template.Spec.Volumes) == 0 {
		return errors.New("there are *no* volumes in this DeploymentConfig to remove")
	}

	found := false

	// If for some reason there is only one volume, let's slice the array to zero length
	// or else you will get a "runtime error: slice bounds of of range [2:1] error
	if len(dc.Spec.Template.Spec.Volumes) == 1 && vol == dc.Spec.Template.Spec.Volumes[0].Name {
		// Mark as found and slice to zero length
		found = true
		dc.Spec.Template.Spec.Volumes = dc.Spec.Template.Spec.Volumes[:0]
	} else {

		for i, volume := range dc.Spec.Template.Spec.Volumes {

			// If we find a match
			if volume.Name == vol {
				found = true

				// Copy (it takes longer, but maintains volume order)
				copy(dc.Spec.Template.Spec.Volumes[i:], dc.Spec.Template.Spec.Volumes[i+1:])
				dc.Spec.Template.Spec.Volumes = dc.Spec.Template.Spec.Volumes[:len(dc.Spec.Template.Spec.Volumes)-1]

				break
			}

		}
	}

	if found {
		return nil
	}

	return fmt.Errorf("unable to find volume '%s' within DeploymentConfig '%s'", vol, dc.ObjectMeta.Name)
}

// removeVolumeMountsFromDC removes the volumeMounts from all the given containers
// in the given Deployment Config and return true. If any of the volumeMount with the name
// is not found, it returns false.
func removeVolumeMountsFromDC(volumeMount string, dc *appsv1.DeploymentConfig) error {

	if len(dc.Spec.Template.Spec.Containers) == 0 {
		return errors.New("something went wrong, there are *no* containers available to iterate through")
	}

	found := false

	for i, container := range dc.Spec.Template.Spec.Containers {

		if len(dc.Spec.Template.Spec.Containers[i].VolumeMounts) == 1 && dc.Spec.Template.Spec.Containers[i].VolumeMounts[0].Name == volumeMount {
			// Mark as found and slice to zero length
			found = true
			dc.Spec.Template.Spec.Containers[i].VolumeMounts = dc.Spec.Template.Spec.Containers[i].VolumeMounts[:0]
		} else {

			for j, volMount := range container.VolumeMounts {

				// If we find a match
				if volMount.Name == volumeMount {
					found = true

					// Copy (it takes longer, but maintains volume mount order)
					copy(dc.Spec.Template.Spec.Containers[i].VolumeMounts[j:], dc.Spec.Template.Spec.Containers[i].VolumeMounts[j+1:])
					dc.Spec.Template.Spec.Containers[i].VolumeMounts = dc.Spec.Template.Spec.Containers[i].VolumeMounts[:len(dc.Spec.Template.Spec.Containers[i].VolumeMounts)-1]

					break
				}
			}
		}
	}

	if found {
		return nil
	}

	return fmt.Errorf("unable to find volume mount '%s'", volumeMount)
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
