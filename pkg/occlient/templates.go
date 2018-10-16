package occlient

import (
	"fmt"

	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CommonImageMeta has all the most common image data that is passed around within Odo
type CommonImageMeta struct {
	Name      string
	Tag       string
	Namespace string
	Ports     []corev1.ContainerPort
}

func generateSupervisordDeploymentConfig(commonObjectMeta metav1.ObjectMeta, builderImage string, commonImageMeta CommonImageMeta, envVar []corev1.EnvVar) appsv1.DeploymentConfig {

	// Generates and deploys a DeploymentConfig with an InitContainer to copy over the SupervisorD binary.
	return appsv1.DeploymentConfig{
		ObjectMeta: commonObjectMeta,
		Spec: appsv1.DeploymentConfigSpec{
			Replicas: 1,
			Selector: map[string]string{
				"deploymentconfig": commonObjectMeta.Name,
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"deploymentconfig": commonObjectMeta.Name,
					},
					// https://github.com/redhat-developer/odo/pull/622#issuecomment-413410736
					Annotations: map[string]string{
						"alpha.image.policy.openshift.io/resolve-names": "*",
					},
				},
				Spec: corev1.PodSpec{
					// The application container
					Containers: []corev1.Container{
						{
							Image: builderImage,
							Name:  commonObjectMeta.Name,
							Ports: commonImageMeta.Ports,
							// Run the actual supervisord binary that has been mounted into the container
							Command: []string{
								"/var/lib/supervisord/bin/supervisord",
							},
							// Using the appropriate configuration file that contains the "run" script for the component.
							// either from: /usr/libexec/s2i/assemble or /opt/app-root/src/.s2i/bin/assemble
							Args: []string{
								"-c",
								"/var/lib/supervisord/conf/supervisor.conf",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      supervisordVolumeName,
									MountPath: "/var/lib/supervisord",
								},
							},
							Env: envVar,
						},
					},

					// Create a volume that will be shared betwen InitContainer and the applicationContainer
					// in order to pass over the SupervisorD binary
					Volumes: []corev1.Volume{
						{
							Name: supervisordVolumeName,
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
			// We provide triggers to create an ImageStream so that the application container will use the
			// correct and approriate image that's located internally within the OpenShift commonObjectMeta.Namespace
			Triggers: []appsv1.DeploymentTriggerPolicy{
				{
					Type: "ConfigChange",
				},
				{
					Type: "ImageChange",
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							commonObjectMeta.Name,
							"copy-files-to-volume",
						},
						From: corev1.ObjectReference{
							Kind:      "ImageStreamTag",
							Name:      fmt.Sprintf("%s:%s", commonImageMeta.Name, commonImageMeta.Tag),
							Namespace: commonImageMeta.Namespace,
						},
					},
				},
			},
		},
	}
}

func generateGitDeploymentConfig(commonObjectMeta metav1.ObjectMeta, image string, containerPorts []corev1.ContainerPort, envVars []corev1.EnvVar) appsv1.DeploymentConfig {
	return appsv1.DeploymentConfig{
		ObjectMeta: commonObjectMeta,
		Spec: appsv1.DeploymentConfigSpec{
			Replicas: 1,
			Selector: map[string]string{
				"deploymentconfig": commonObjectMeta.Name,
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"deploymentconfig": commonObjectMeta.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: image,
							Name:  commonObjectMeta.Name,
							Ports: containerPorts,
							Env:   envVars,
						},
					},
				},
			},
			Triggers: []appsv1.DeploymentTriggerPolicy{
				{
					Type: "ConfigChange",
				},
				{
					Type: "ImageChange",
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							commonObjectMeta.Name,
						},
						From: corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: image,
						},
					},
				},
			},
		},
	}
}

// generateBuildConfig creates a BuildConfig for Git URL's being passed into Odo
func generateBuildConfig(commonObjectMeta metav1.ObjectMeta, gitURL string, imageName string, imageNamespace string) buildv1.BuildConfig {

	buildSource := buildv1.BuildSource{
		Git: &buildv1.GitBuildSource{
			URI: gitURL,
		},
		Type: buildv1.BuildSourceGit,
	}

	return buildv1.BuildConfig{
		ObjectMeta: commonObjectMeta,
		Spec: buildv1.BuildConfigSpec{
			CommonSpec: buildv1.CommonSpec{
				Output: buildv1.BuildOutput{
					To: &corev1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: commonObjectMeta.Name + ":latest",
					},
				},
				Source: buildSource,
				Strategy: buildv1.BuildStrategy{
					SourceStrategy: &buildv1.SourceBuildStrategy{
						From: corev1.ObjectReference{
							Kind:      "ImageStreamTag",
							Name:      imageName,
							Namespace: imageNamespace,
						},
					},
				},
			},
		},
	}
}

//
// Below is related to SUPERVISORD
//

// AddBootstrapInitContainer adds the bootstrap init container to the deployment config
// dc is the deployment config to be updated
// dcName is the name of the deployment config
func addBootstrapVolumeCopyInitContainer(dc *appsv1.DeploymentConfig, dcName string) {
	dc.Spec.Template.Spec.InitContainers = append(dc.Spec.Template.Spec.InitContainers,
		corev1.Container{
			Name: "copy-files-to-volume",
			// Using custom image from bootstrapperImage variable for the initial initContainer
			Image: dc.Spec.Template.Spec.Containers[0].Image,
			Command: []string{
				"sh",
				"-c"},
			// Script required to copy over file information from /opt/app-root
			// Source https://github.com/jupyter-on-openshift/jupyter-notebooks/blob/master/minimal-notebook/setup-volume.sh
			Args: []string{`
SRC=/opt/app-root
DEST=/mnt/app-root

if [ -f $DEST/.delete-volume ]; then
    rm -rf $DEST
fi
 if [ -d $DEST ]; then
    if [ -f $DEST/.sync-volume ]; then
        if ! [[ "$JUPYTER_SYNC_VOLUME" =~ ^(false|no|n|0)$ ]]; then
            JUPYTER_SYNC_VOLUME=yes
        fi
    fi
     if [[ "$JUPYTER_SYNC_VOLUME" =~ ^(true|yes|y|1)$ ]]; then
        rsync -ar --ignore-existing $SRC/. $DEST
    fi
     exit
fi
 if [ -d $DEST.setup-volume ]; then
    rm -rf $DEST.setup-volume
fi

mkdir -p $DEST.setup-volume
tar -C $SRC -cf - . | tar -C $DEST.setup-volume -xvf -
mv $DEST.setup-volume $DEST
			`},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      getAppRootVolumeName(dcName),
					MountPath: "/mnt",
				},
			},
		})
}

// addBootstrapSupervisordInitContainer creates an init container that will copy over
// supervisord to the application image during the start-up procress.
func addBootstrapSupervisordInitContainer(dc *appsv1.DeploymentConfig, dcName string) {

	dc.Spec.Template.Spec.InitContainers = append(dc.Spec.Template.Spec.InitContainers,
		corev1.Container{
			Name:  "copy-supervisord",
			Image: bootstrapperImage,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      supervisordVolumeName,
					MountPath: "/var/lib/supervisord",
				},
			},
			Command: []string{
				"/usr/bin/cp",
			},
			Args: []string{
				"-r",
				"/opt/supervisord",
				"/var/lib/",
			},
		})
}

// addBootstrapVolume adds the bootstrap volume to the deployment config
// dc is the deployment config to be updated
// dcName is the name of the deployment config
func addBootstrapVolume(dc *appsv1.DeploymentConfig, dcName string) {
	dc.Spec.Template.Spec.Volumes = append(dc.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: getAppRootVolumeName(dcName),
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: getAppRootVolumeName(dcName),
			},
		},
	})
}

// addBootstrapVolumeMount mounts the bootstrap volume to the deployment config
// dc is the deployment config to be updated
// dcName is the name of the deployment config
func addBootstrapVolumeMount(dc *appsv1.DeploymentConfig, dcName string) {
	for i := range dc.Spec.Template.Spec.Containers {
		dc.Spec.Template.Spec.Containers[i].VolumeMounts = append(dc.Spec.Template.Spec.Containers[i].VolumeMounts, corev1.VolumeMount{
			Name:      getAppRootVolumeName(dcName),
			MountPath: "/opt/app-root",
			SubPath:   "app-root",
		})
	}
}
