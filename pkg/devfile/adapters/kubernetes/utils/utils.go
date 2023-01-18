package utils

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfileParser "github.com/devfile/library/v2/pkg/devfile/parser"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/storage"
)

const (
	// _envProjectsRoot is the env defined for project mount in a component container when component's mountSources=true
	_envProjectsRoot = "PROJECTS_ROOT"
)

// GetOdoContainerVolumes returns the mandatory Kube volumes for an Odo component
func GetOdoContainerVolumes(sourcePVCName string) []corev1.Volume {
	var sourceVolume corev1.Volume

	if sourcePVCName != "" {
		sourceVolume = corev1.Volume{
			Name: storage.OdoSourceVolume,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: sourcePVCName},
			},
		}
	} else {
		sourceVolume = corev1.Volume{
			Name: storage.OdoSourceVolume,
		}
	}

	return []corev1.Volume{
		sourceVolume,
		{
			// Create a volume that will be shared between InitContainer and the applicationContainer
			// in order to pass over some files for odo
			Name: storage.SharedDataVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
}

// AddOdoProjectVolume adds the odo project volume to the containers
func AddOdoProjectVolume(containers *[]corev1.Container) {
	if containers == nil {
		return
	}
	for i, container := range *containers {
		for _, env := range container.Env {
			if env.Name == _envProjectsRoot {
				container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
					Name:      storage.OdoSourceVolume,
					MountPath: env.Value,
				})
				(*containers)[i] = container
			}
		}
	}
}

// AddOdoMandatoryVolume adds the odo mandatory volumes to the containers
func AddOdoMandatoryVolume(containers *[]corev1.Container) {
	if containers == nil {
		return
	}
	for i, container := range *containers {
		klog.V(2).Infof("Updating container %v with mandatory volume mounts", container.Name)
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
			Name:      storage.SharedDataVolumeName,
			MountPath: storage.SharedDataMountPath,
		})
		(*containers)[i] = container
	}
}

// UpdateContainersEntrypointsIfNeeded updates the run components entrypoint
// if no entrypoint has been specified for the component in the devfile
func UpdateContainersEntrypointsIfNeeded(
	devfileObj devfileParser.DevfileObj,
	containers []corev1.Container,
	devfileBuildCmd string,
	devfileRunCmd string,
	devfileDebugCmd string,
) ([]corev1.Container, error) {
	buildCommand, hasBuildCmd, err := libdevfile.GetCommand(devfileObj, devfileBuildCmd, v1alpha2.BuildCommandGroupKind)
	if err != nil {
		return nil, err
	}
	runCommand, hasRunCmd, err := libdevfile.GetCommand(devfileObj, devfileRunCmd, v1alpha2.RunCommandGroupKind)
	if err != nil {
		return nil, err
	}
	debugCommand, hasDebugCmd, err := libdevfile.GetCommand(devfileObj, devfileDebugCmd, v1alpha2.DebugCommandGroupKind)
	if err != nil {
		return nil, err
	}

	var cmdsToHandle []v1alpha2.Command
	if hasBuildCmd {
		cmdsToHandle = append(cmdsToHandle, buildCommand)
	}
	if hasRunCmd {
		cmdsToHandle = append(cmdsToHandle, runCommand)
	}
	if hasDebugCmd {
		cmdsToHandle = append(cmdsToHandle, debugCommand)
	}

	var components []string
	var containerComps []string
	for _, cmd := range cmdsToHandle {
		containerComps, err = libdevfile.GetContainerComponentsForCommand(devfileObj, cmd)
		if err != nil {
			return nil, err
		}
		components = append(components, containerComps...)
	}

	for _, c := range components {
		for i := range containers {
			container := &containers[i]
			if container.Name == c {
				overrideContainerCommandAndArgsIfNeeded(container)
			}
		}
	}

	return containers, nil

}

// overrideContainerCommandAndArgsIfNeeded overrides the container's entrypoint
// if the corresponding component does not have any command and/or args in the Devfile.
// This is a workaround until the default Devfile registry exposes stacks with non-terminating containers.
func overrideContainerCommandAndArgsIfNeeded(container *corev1.Container) {
	if len(container.Command) != 0 || len(container.Args) != 0 {
		return
	}

	klog.V(2).Infof("No entrypoint defined for the component, setting container %v entrypoint to 'tail -f /dev/null'. You can set a 'command' and/or 'args' for the component to override this default entrypoint.", container.Name)
	// #5768: overriding command and args if the container had no Command or Args defined in it.
	// This is a workaround for us to quickly switch to running without Supervisord,
	// while waiting for the Devfile registry to expose stacks with non-terminating containers.
	// TODO(rm3l): Remove this once https://github.com/devfile/registry/pull/102 is merged on the Devfile side
	container.Command = []string{"tail"}
	container.Args = []string{"-f", "/dev/null"}
}
