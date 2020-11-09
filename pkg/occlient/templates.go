package occlient

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/kclient/generator"

	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// appRootSubPath defines the sup-path in the odo's PV where app root will recide
	appRootSubPath = "app-root"
	// deploymentDirSubPath defines the sup-path in the odo's PV where the deployment dir will recide
	deploymentDirSubPath = "deployment"
)

// CommonImageMeta has all the most common image data that is passed around within Odo
type CommonImageMeta struct {
	Name      string
	Tag       string
	Namespace string
	Ports     []corev1.ContainerPort
}

type SupervisorDUpdateParams struct {
	existingDc           *appsv1.DeploymentConfig
	commonObjectMeta     metav1.ObjectMeta
	commonImageMeta      CommonImageMeta
	envVar               []corev1.EnvVar
	envFrom              []corev1.EnvFromSource
	resourceRequirements *corev1.ResourceRequirements
}

// generateSupervisordDeploymentConfig generates dc for local and binary components
// Parameters:
//	commonObjectMeta: Contains annotations and labels for dc
//	commonImageMeta: Contains details like image NS, name, tag and ports to be exposed
//	envVar: env vars to be exposed
//	resourceRequirements: Container cpu and memory resource requirements
// Returns:
//	deployment config generated using above parameters
func generateSupervisordDeploymentConfig(commonObjectMeta metav1.ObjectMeta, commonImageMeta CommonImageMeta,
	envVar []corev1.EnvVar, envFrom []corev1.EnvFromSource, resourceRequirements *corev1.ResourceRequirements) appsv1.DeploymentConfig {

	if commonImageMeta.Namespace == "" {
		commonImageMeta.Namespace = "openshift"
	}

	// Generates and deploys a DeploymentConfig with an InitContainer to copy over the SupervisorD binary.
	dc := appsv1.DeploymentConfig{
		ObjectMeta: commonObjectMeta,
		Spec: appsv1.DeploymentConfigSpec{
			Replicas: 1,
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.DeploymentStrategyTypeRecreate,
			},
			Selector: map[string]string{
				"deploymentconfig": commonObjectMeta.Name,
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"deploymentconfig": commonObjectMeta.Name,
					},
					// https://github.com/openshift/odo/pull/622#issuecomment-413410736
					Annotations: map[string]string{
						"alpha.image.policy.openshift.io/resolve-names": "*",
					},
				},
				Spec: corev1.PodSpec{
					// The application container
					Containers: []corev1.Container{
						{
							Image: " ",
							Name:  commonObjectMeta.Name,
							Ports: commonImageMeta.Ports,
							// Run the actual supervisord binary that has been mounted into the container
							Command: []string{
								"/opt/odo/bin/go-init",
							},
							// Using the appropriate configuration file that contains the "run" script for the component.
							// either from: /usr/libexec/s2i/assemble or /opt/app-root/src/.s2i/bin/assemble
							Args: []string{
								"-pre",
								"/opt/odo/bin/s2i-setup",
								"-main",
								"/opt/odo/bin/supervisord -c /opt/app-root/conf/supervisor.conf",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      common.SupervisordVolumeName,
									MountPath: "/opt/odo/",
								},
							},
							Env:     envVar,
							EnvFrom: envFrom,
						},
					},

					// Create a volume that will be shared between InitContainer and the applicationContainer
					// in order to pass over the SupervisorD binary
					Volumes: []corev1.Volume{
						{
							Name: common.SupervisordVolumeName,
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
			// We provide triggers to create an ImageStream so that the application container will use the
			// correct and appropriate image that's located internally within the OpenShift commonObjectMeta.Namespace
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
	containerIndex := -1
	if resourceRequirements != nil {
		for index, container := range dc.Spec.Template.Spec.Containers {
			if container.Name == commonObjectMeta.Name {
				containerIndex = index
				break
			}
		}
		if containerIndex != -1 {
			dc.Spec.Template.Spec.Containers[containerIndex].Resources = *resourceRequirements
		}
	}
	return dc
}

// updateSupervisorDeploymentConfig updates the deploymentConfig during push
// updateParams are the parameters used during the update
func updateSupervisorDeploymentConfig(updateParams SupervisorDUpdateParams) appsv1.DeploymentConfig {

	dc := *updateParams.existingDc

	dc.ObjectMeta = updateParams.commonObjectMeta

	if len(dc.Spec.Template.Spec.Containers) > 0 {
		dc.Spec.Template.Spec.Containers[0].Name = updateParams.commonObjectMeta.Name
		dc.Spec.Template.Spec.Containers[0].Ports = updateParams.commonImageMeta.Ports
		dc.Spec.Template.Spec.Containers[0].Env = updateParams.envVar
		dc.Spec.Template.Spec.Containers[0].EnvFrom = updateParams.envFrom

		if updateParams.resourceRequirements != nil && dc.Spec.Template.Spec.Containers[0].Name == updateParams.commonObjectMeta.Name {
			dc.Spec.Template.Spec.Containers[0].Resources = *updateParams.resourceRequirements
		}
	}

	return dc
}

// FetchContainerResourceLimits returns cpu and memory resource limits of the component container from the passed dc
// Parameter:
//	container: Component container
// Returns:
//	resource limits from passed component container
func FetchContainerResourceLimits(container corev1.Container) corev1.ResourceRequirements {
	return container.Resources
}

// parseResourceQuantity takes a string representation of quantity/amount of a resource and returns kubernetes representation of it and errors if any
// This is a wrapper around the kube client provided ParseQuantity added to in future support more units and make it more readable
func parseResourceQuantity(resQuantity string) (resource.Quantity, error) {
	return resource.ParseQuantity(resQuantity)
}

// GetResourceRequirementsFromCmpSettings converts the cpu and memory request info from component configuration into format usable in dc
// Parameters:
//	cfg: Compoennt configuration/settings
// Returns:
//	*corev1.ResourceRequirements: component configuration converted into format usable in dc
func GetResourceRequirementsFromCmpSettings(cfg config.LocalConfigInfo) (*corev1.ResourceRequirements, error) {
	var resourceRequirements corev1.ResourceRequirements
	requests := make(corev1.ResourceList)
	limits := make(corev1.ResourceList)

	cfgMinCPU := cfg.GetMinCPU()
	cfgMaxCPU := cfg.GetMaxCPU()
	cfgMinMemory := cfg.GetMinMemory()
	cfgMaxMemory := cfg.GetMaxMemory()

	if cfgMinCPU != "" {
		minCPU, err := parseResourceQuantity(cfgMinCPU)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse the min cpu")
		}
		requests[corev1.ResourceCPU] = minCPU
	}

	if cfgMaxCPU != "" {
		maxCPU, err := parseResourceQuantity(cfgMaxCPU)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse max cpu")
		}
		limits[corev1.ResourceCPU] = maxCPU
	}

	if cfgMinMemory != "" {
		minMemory, err := parseResourceQuantity(cfgMinMemory)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse min memory")
		}
		requests[corev1.ResourceMemory] = minMemory
	}

	if cfgMaxMemory != "" {
		maxMemory, err := parseResourceQuantity(cfgMaxMemory)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse max memory")
		}
		limits[corev1.ResourceMemory] = maxMemory
	}

	if len(limits) > 0 {
		resourceRequirements.Limits = limits
	}

	if len(requests) > 0 {
		resourceRequirements.Requests = requests
	}

	return &resourceRequirements, nil
}

func generateGitDeploymentConfig(commonObjectMeta metav1.ObjectMeta, image string, containerPorts []corev1.ContainerPort, envVars []corev1.EnvVar, resourceRequirements *corev1.ResourceRequirements) appsv1.DeploymentConfig {
	dc := appsv1.DeploymentConfig{
		ObjectMeta: commonObjectMeta,
		Spec: appsv1.DeploymentConfigSpec{
			Replicas: 1,
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.DeploymentStrategyTypeRecreate,
			},
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
							// image stream trigger and specifying an image at the same time are mutually exclusive options
							// thus we put "" into image field as we are specifying an image stream trigger
							Image: "",
							Name:  commonObjectMeta.Name,
							Ports: containerPorts,
							Env:   envVars,
						},
					},
				},
			},
			Triggers: []appsv1.DeploymentTriggerPolicy{
				{
					Type: "ImageChange",
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						// setting automatic to false so that the trigger is disabled and a new image doesn't trigger deployment
						// we don't remove this trigger so that we don't face image resolution issues
						Automatic: false,
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
	containerIndex := -1
	if resourceRequirements != nil {
		for index, container := range dc.Spec.Template.Spec.Containers {
			if container.Name == commonObjectMeta.Name {
				containerIndex = index
				break
			}
		}
		if containerIndex != -1 {
			dc.Spec.Template.Spec.Containers[containerIndex].Resources = *resourceRequirements
		}
	}
	return dc
}

// generateBuildConfig creates a BuildConfig for Git URL's being passed into Odo
func generateBuildConfig(commonObjectMeta metav1.ObjectMeta, gitURL, gitRef, imageName, imageNamespace string) buildv1.BuildConfig {

	params := generator.BuildConfigParams{
		CommonObjectMeta: commonObjectMeta,
		GitURL:           gitURL,
		GitRef:           gitRef,
		BuildStrategy:    generator.GetSourceBuildStrategy(imageName, imageNamespace),
	}

	return generator.GetBuildConfig(params)
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
// supervisord to the application image during the start-up process.
func addBootstrapSupervisordInitContainer(dc *appsv1.DeploymentConfig, dcName string) {

	dc.Spec.Template.Spec.InitContainers = append(dc.Spec.Template.Spec.InitContainers,
		corev1.Container{
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
	addVolumeMount(dc, getAppRootVolumeName(dcName), DefaultAppRootDir, appRootSubPath)
}

// addDeploymentDirVolumeMount mounts the bootstrap volume to the deployment config
// in a sub path where the ODO_S2I_DEPLOYMENT_DIR is present for optimisation purposes
// dc is the deployment config to be updated
func addDeploymentDirVolumeMount(dc *appsv1.DeploymentConfig, mountPath string) {
	addVolumeMount(dc, getAppRootVolumeName(dc.Name), mountPath, deploymentDirSubPath)
}

// addVolumeMount adds a volume mount to the deployment config
// dc is the deployment config to be updated
func addVolumeMount(dc *appsv1.DeploymentConfig, name, mountPath, subPath string) {
	for i := range dc.Spec.Template.Spec.Containers {
		dc.Spec.Template.Spec.Containers[i].VolumeMounts = append(dc.Spec.Template.Spec.Containers[i].VolumeMounts, corev1.VolumeMount{
			Name:      name,
			MountPath: mountPath,
			SubPath:   subPath,
		})
	}
}
