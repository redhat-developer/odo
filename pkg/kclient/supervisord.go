package kclient

import corev1 "k8s.io/api/core/v1"

// AddBootstrapSupervisordInitContainer creates an init container that will copy over
// supervisord to the application image during the start-up procress.
func AddBootstrapSupervisordInitContainer(podTemplateSpec *corev1.PodTemplateSpec) {

	podTemplateSpec.Spec.InitContainers = append(podTemplateSpec.Spec.InitContainers,
		corev1.Container{
			Name:  supervisordInitContainerName,
			Image: defaultBootstrapperImage,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      GetSupervisordVolumeName(),
					MountPath: "/opt/odo/",
				},
			},
			Command: []string{
				"/usr/bin/cp",
			},
			Args: []string{
				"-r",
				"/opt/odo-init/.",
				"/opt/odo/",
			},
		})
}
