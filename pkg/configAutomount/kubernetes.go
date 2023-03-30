package configAutomount

import (
	"path/filepath"

	"github.com/redhat-developer/odo/pkg/kclient"
)

const (
	labelMountName  = "controller.devfile.io/mount-to-devworkspace"
	labelMountValue = "true"

	annotationMountPathName = "controller.devfile.io/mount-path"
)

type KubernetesClient struct {
	kubeClient kclient.ClientInterface
}

func NewKubernetesClient(kubeClient kclient.ClientInterface) KubernetesClient {
	return KubernetesClient{
		kubeClient: kubeClient,
	}
}

func (o KubernetesClient) GetAutomountingVolumes() ([]AutomountInfo, error) {
	var result []AutomountInfo

	pvcs, err := o.getAutomountingPVCs()
	if err != nil {
		return nil, err
	}
	result = append(result, pvcs...)

	secrets, err := o.getAutomountingSecrets()
	if err != nil {
		return nil, err
	}
	result = append(result, secrets...)

	cms, err := o.getAutomountingConfigmaps()
	if err != nil {
		return nil, err
	}
	result = append(result, cms...)

	return result, nil
}

func (o KubernetesClient) getAutomountingPVCs() ([]AutomountInfo, error) {
	pvcs, err := o.kubeClient.ListPVCs(labelMountName + "=" + labelMountValue)
	if err != nil {
		return nil, err
	}

	var result []AutomountInfo
	for _, pvc := range pvcs {
		mountPath := filepath.ToSlash(filepath.Join("/", "tmp", pvc.Name))
		if val, found := getMountPathFromAnnotation(pvc.Annotations); found {
			mountPath = val
		}
		result = append(result, AutomountInfo{
			VolumeType: VolumeTypePVC,
			VolumeName: pvc.Name,
			MountPath:  mountPath,
			MountAs:    MountAsFile,
			ReadOnly:   false, // TODO consider annotation "controller.devfile.io/read-only"
		})
	}
	return result, nil
}

func (o KubernetesClient) getAutomountingSecrets() ([]AutomountInfo, error) {
	secrets, err := o.kubeClient.ListSecrets(labelMountName + "=" + labelMountValue)
	if err != nil {
		return nil, err
	}

	var result []AutomountInfo
	for _, secret := range secrets {
		mountPath := filepath.ToSlash(filepath.Join("/", "etc", "secret", secret.Name))
		if val, found := getMountPathFromAnnotation(secret.Annotations); found {
			mountPath = val
		}
		result = append(result, AutomountInfo{
			VolumeType: VolumeTypeSecret,
			VolumeName: secret.Name,
			MountPath:  mountPath,
			MountAs:    MountAsFile, // TODO consider annotation "controller.devfile.io/mount-as"
			ReadOnly:   false,       // TODO consider annotation "controller.devfile.io/read-only"
		})
	}
	return result, nil
}

func (o KubernetesClient) getAutomountingConfigmaps() ([]AutomountInfo, error) {
	cms, err := o.kubeClient.ListConfigMaps(labelMountName + "=" + labelMountValue)
	if err != nil {
		return nil, err
	}

	var result []AutomountInfo
	for _, cm := range cms {
		mountPath := filepath.ToSlash(filepath.Join("/", "etc", "config", cm.Name))
		if val, found := getMountPathFromAnnotation(cm.Annotations); found {
			mountPath = val
		}
		result = append(result, AutomountInfo{
			VolumeType: VolumeTypeConfigmap,
			VolumeName: cm.Name,
			MountPath:  mountPath,
			MountAs:    MountAsFile, // TODO consider annotation "controller.devfile.io/mount-as"
			ReadOnly:   false,       // TODO consider annotation "controller.devfile.io/read-only"
		})
	}
	return result, nil
}

func getMountPathFromAnnotation(annotations map[string]string) (string, bool) {
	val, found := annotations[annotationMountPathName]
	return val, found
}
