package kclient

import corev1 "k8s.io/api/core/v1"

const (
	// The init container name for supervisord
	supervisordInitContainerName = "copy-supervisord"

	// The default image for odo init bootstrapper container
	defaultBootstrapperImage = "jeevandroid/odo-init-image"

	// Create a custom name and (hope) that users don't use the *exact* same name in their deployment (occlient.go)
	supervisordVolumeName = "odo-supervisord-shared-data"

	// The supervisord Mount Path for the container mounting the supervisord volume
	supervisordMountPath = "/opt/odo/"

	// The supervisord binary path inside the container volume mount
	supervisordBinaryPath = "/opt/odo/bin/supervisord"

	// The supervisord configuration file inside the container volume mount
	supervisordConfFile = "/opt/odo/conf/devfile-supervisor.conf"

	// The path to the odo init image contents
	odoInitImageContents = "/opt/odo-init/."
)

// GetSupervisordVolumeName returns the supervisord Volume Name
func GetSupervisordVolumeName() string {
	return supervisordVolumeName
}

// GetSupervisordMountPath returns the supervisord Volume Name
func GetSupervisordMountPath() string {
	return supervisordMountPath
}

// GetSupervisordBinaryPath returns the path to the supervisord binary
func GetSupervisordBinaryPath() string {
	return supervisordBinaryPath
}

// GetSupervisordConfFilePath returns the path to the supervisord conf file
func GetSupervisordConfFilePath() string {
	return supervisordConfFile
}

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
					MountPath: supervisordMountPath,
				},
			},
			Command: []string{
				"/usr/bin/cp",
			},
			Args: []string{
				"-r",
				odoInitImageContents,
				supervisordMountPath,
			},
		})
}
