package kclient

import (
	"github.com/openshift/odo/pkg/devfile/adapters/common"

	corev1 "k8s.io/api/core/v1"
)

// AddBootstrapSupervisordInitContainer creates an init container that will copy over
// supervisord to the application image during the start-up procress.
func AddBootstrapSupervisordInitContainer() corev1.Container {

	return corev1.Container{
		Name:  common.SupervisordInitContainerName,
		Image: common.GetBootstrapperImage(),
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      common.SupervisordVolumeName,
				MountPath: common.SupervisordMountPath,
			},
		},
		Command: []string{
			"/usr/bin/cp",
		},
		Args: []string{
			"-r",
			common.OdoInitImageContents,
			common.SupervisordMountPath,
		},
	}
}
