package utils

import (
	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/util"
	"strconv"

	"github.com/openshift/odo/pkg/storage"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
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
			// in order to pass over the SupervisorD binary
			Name: adaptersCommon.SupervisordVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
}

// ComponentExists checks whether a deployment by the given name exists in the given app
func ComponentExists(client kclient.Client, name string, app string) (bool, error) {
	deployment, err := client.GetOneDeployment(name, app)
	if _, ok := err.(*kclient.DeploymentNotFoundError); ok {
		klog.V(2).Infof("Deployment %s not found for belonging to the %s app ", name, app)
		return false, nil
	}
	return deployment != nil, err
}

// isEnvPresent checks if the env variable is present in an array of env variables
func isEnvPresent(EnvVars []corev1.EnvVar, envVarName string) bool {
	isPresent := false

	for _, envVar := range EnvVars {
		if envVar.Name == envVarName {
			isPresent = true
		}
	}

	return isPresent
}

// AddOdoProjectVolume adds the odo project volume to the containers
func AddOdoProjectVolume(containers *[]corev1.Container) {
	for i, container := range *containers {
		for _, env := range container.Env {
			if env.Name == adaptersCommon.EnvProjectsRoot {
				container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
					Name:      storage.OdoSourceVolume,
					MountPath: env.Value,
				})
				(*containers)[i] = container
			}
		}
	}
}

// UpdateContainersWithSupervisord updates the run components entrypoint and volume mount
// with supervisord if no entrypoint has been specified for the component in the devfile
func UpdateContainersWithSupervisord(devfileObj devfileParser.DevfileObj, containers []corev1.Container, devfileRunCmd string, devfileDebugCmd string, devfileDebugPort int) ([]corev1.Container, error) {

	runCommand, err := adaptersCommon.GetRunCommand(devfileObj.Data, devfileRunCmd)
	if err != nil {
		return nil, err
	}

	debugCommand, err := adaptersCommon.GetDebugCommand(devfileObj.Data, devfileDebugCmd)
	if err != nil {
		return nil, err
	}

	for i := range containers {
		container := &containers[i]
		// Check if the container belongs to a run command component
		if container.Name == runCommand.Exec.Component {
			// If the run component container has no entrypoint and arguments, override the entrypoint with supervisord
			if len(container.Command) == 0 && len(container.Args) == 0 {
				overrideContainerArgs(container)
			}

			// Always mount the supervisord volume in the run component container
			klog.V(2).Infof("Updating container %v with supervisord volume mounts", container.Name)
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				Name:      adaptersCommon.SupervisordVolumeName,
				MountPath: adaptersCommon.SupervisordMountPath,
			})

			// Update the run container's ENV for work dir and command
			// only if the env var is not set in the devfile
			// This is done, so supervisord can use it in it's program
			if !isEnvPresent(container.Env, adaptersCommon.EnvOdoCommandRun) {
				klog.V(2).Infof("Updating container %v env with run command", container.Name)
				var command string
				setEnvVariable := util.GetCommandStringFromEnvs(runCommand.Exec.Env)

				if setEnvVariable == "" {
					command = runCommand.Exec.CommandLine
				} else {
					command = setEnvVariable + " && " + runCommand.Exec.CommandLine
				}
				container.Env = append(container.Env,
					corev1.EnvVar{
						Name:  adaptersCommon.EnvOdoCommandRun,
						Value: command,
					})
			}

			if !isEnvPresent(container.Env, adaptersCommon.EnvOdoCommandRunWorkingDir) && runCommand.Exec.WorkingDir != "" {
				klog.V(2).Infof("Updating container %v env with run command's workdir", container.Name)
				container.Env = append(container.Env,
					corev1.EnvVar{
						Name:  adaptersCommon.EnvOdoCommandRunWorkingDir,
						Value: runCommand.Exec.WorkingDir,
					})
			}
		}

		// Check if the container belongs to a debug command component
		if debugCommand.Exec != nil && container.Name == debugCommand.Exec.Component {
			// If the debug component container has no entrypoint and arguments, override the entrypoint with supervisord
			if len(container.Command) == 0 && len(container.Args) == 0 {
				overrideContainerArgs(container)
			}

			foundMountPath := false
			for _, mounts := range container.VolumeMounts {
				if mounts.Name == adaptersCommon.SupervisordVolumeName && mounts.MountPath == adaptersCommon.SupervisordMountPath {
					foundMountPath = true
				}
			}

			if !foundMountPath {
				// Always mount the supervisord volume in the debug component container
				klog.V(2).Infof("Updating container %v with supervisord volume mounts", container.Name)
				container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
					Name:      adaptersCommon.SupervisordVolumeName,
					MountPath: adaptersCommon.SupervisordMountPath,
				})
			}

			// Update the debug container's ENV for work dir and command
			// only if the env var is not set in the devfile
			// This is done, so supervisord can use it in it's program
			if !isEnvPresent(container.Env, adaptersCommon.EnvOdoCommandDebug) {
				klog.V(2).Infof("Updating container %v env with debug command", container.Name)
				var command string
				setEnvVariable := util.GetCommandStringFromEnvs(debugCommand.Exec.Env)

				if setEnvVariable == "" {
					command = debugCommand.Exec.CommandLine
				} else {
					command = setEnvVariable + " && " + debugCommand.Exec.CommandLine
				}
				container.Env = append(container.Env,
					corev1.EnvVar{
						Name:  adaptersCommon.EnvOdoCommandDebug,
						Value: command,
					})
			}

			if debugCommand.Exec.WorkingDir != "" && !isEnvPresent(container.Env, adaptersCommon.EnvOdoCommandDebugWorkingDir) {
				klog.V(2).Infof("Updating container %v env with debug command's workdir", container.Name)
				container.Env = append(container.Env,
					corev1.EnvVar{
						Name:  adaptersCommon.EnvOdoCommandDebugWorkingDir,
						Value: debugCommand.Exec.WorkingDir,
					})
			}

			if !isEnvPresent(container.Env, adaptersCommon.EnvDebugPort) {
				klog.V(2).Infof("Updating container %v env with debug command's debugPort", container.Name)
				container.Env = append(container.Env,
					corev1.EnvVar{
						Name:  adaptersCommon.EnvDebugPort,
						Value: strconv.Itoa(devfileDebugPort),
					})
			} else {
				// if the env has the debug port but if its value is different then we change it
				// this way the value of env info is given more importance then whats there on the devfile
				for i, envVar := range container.Env {
					if envVar.Name == adaptersCommon.EnvDebugPort {
						if envVar.ValueFrom == nil && (envVar.Value != strconv.Itoa(devfileDebugPort)) {
							container.Env[i].Value = strconv.Itoa(devfileDebugPort)
							break
						}
					}
				}
			}
		}
	}

	return containers, nil

}

// overrideContainerArgs overrides the container's entrypoint with supervisord
func overrideContainerArgs(container *corev1.Container) {
	klog.V(2).Infof("Updating container %v entrypoint with supervisord", container.Name)
	container.Command = append(container.Command, adaptersCommon.SupervisordBinaryPath)
	container.Args = append(container.Args, "-c", adaptersCommon.SupervisordConfFile)
}
