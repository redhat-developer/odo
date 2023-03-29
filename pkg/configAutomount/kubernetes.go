package configAutomount

import (
	"path/filepath"

	"github.com/redhat-developer/odo/pkg/kclient"
)

const (
	labelMountName  = "controller.devfile.io/mount-to-devworkspace"
	labelMountValue = "true"
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
	pvcs, err := o.kubeClient.ListPVCs(labelMountName + "=" + labelMountValue)
	if err != nil {
		return nil, err
	}
	var result []AutomountInfo
	for _, pvc := range pvcs {
		result = append(result, AutomountInfo{
			VolumeType: VolumeTypePVC,
			VolumeName: pvc.Name,
			MountPath:  filepath.ToSlash(filepath.Join("/", "tmp", pvc.Name)), // TODO consider annotation "controller.devfile.io/mount-path"
			MountAs:    MountAsFile,
			ReadOnly:   false, // TODO consider annotation "controller.devfile.io/read-only"
		})
	}
	return result, nil
}
