package utils

import (
	"fmt"
	"strconv"
	"strings"

	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/kclient/generator"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/util"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
)

const (
	containerNameMaxLen = 55
)

// GetOdoContainerVolumes returns the mandatory Kube volumes for an Odo component
func GetOdoContainerVolumes() []corev1.Volume {
	return []corev1.Volume{
		{
			Name: generator.DevfileSourceVolume,
		},
		{
			// Create a volume that will be shared betwen InitContainer and the applicationContainer
			// in order to pass over the SupervisorD binary
			Name: adaptersCommon.SupervisordVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
}

// ComponentExists checks whether a deployment by the given name exists
func ComponentExists(client kclient.Client, name string) (bool, error) {
	deployment, err := client.GetDeploymentByName(name)
	if kerrors.IsNotFound(err) {
		klog.V(2).Infof("Deployment %s not found", name)
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
				var setEnvVariable, command string
				for _, envVar := range runCommand.Exec.Env {
					setEnvVariable = setEnvVariable + fmt.Sprintf("%v=\"%v\" ", envVar.Name, envVar.Value)
				}
				if setEnvVariable == "" {
					command = runCommand.Exec.CommandLine
				} else {
					command = setEnvVariable + "&& " + runCommand.Exec.CommandLine
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
				var setEnvVariable, command string
				for _, envVar := range debugCommand.Exec.Env {
					setEnvVariable = setEnvVariable + fmt.Sprintf("%v=\"%v\" ", envVar.Name, envVar.Value)
				}
				if setEnvVariable == "" {
					command = debugCommand.Exec.CommandLine
				} else {
					command = setEnvVariable + "&& " + debugCommand.Exec.CommandLine
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
			}
		}
	}

	return containers, nil

}

// // GetResourceReqs creates a kubernetes ResourceRequirements object based on resource requirements set in the devfile
// func GetResourceReqs(comp common.DevfileComponent) corev1.ResourceRequirements {
// 	reqs := corev1.ResourceRequirements{}
// 	limits := make(corev1.ResourceList)
// 	if comp.Container.MemoryLimit != "" {
// 		memoryLimit, err := resource.ParseQuantity(comp.Container.MemoryLimit)
// 		if err == nil {
// 			limits[corev1.ResourceMemory] = memoryLimit
// 		}
// 		reqs.Limits = limits
// 	}
// 	return reqs
// }

// overrideContainerArgs overrides the container's entrypoint with supervisord
func overrideContainerArgs(container *corev1.Container) {
	klog.V(2).Infof("Updating container %v entrypoint with supervisord", container.Name)
	container.Command = append(container.Command, adaptersCommon.SupervisordBinaryPath)
	container.Args = append(container.Args, "-c", adaptersCommon.SupervisordConfFile)
}

// UpdateContainerWithEnvFrom populates the runtime container with relevant
// values for "EnvFrom" so that component can be linked with Operator backed
// service
func UpdateContainerWithEnvFrom(containers []corev1.Container, devfile devfileParser.DevfileObj, devfileRunCmd string, ei envinfo.EnvSpecificInfo) ([]corev1.Container, error) {
	runCommand, err := adaptersCommon.GetRunCommand(devfile.Data, devfileRunCmd)
	if err != nil {
		return nil, err
	}

	for i := range containers {
		c := &containers[i]
		if c.Name == runCommand.Exec.Component {
			c.EnvFrom = generateEnvFromSource(ei)
		}
	}

	return containers, nil
}

func generateEnvFromSource(ei envinfo.EnvSpecificInfo) []corev1.EnvFromSource {

	envFrom := []corev1.EnvFromSource{}

	for _, link := range ei.GetLink() {
		envFrom = append(envFrom, corev1.EnvFromSource{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: link.Name,
				},
			},
		})
	}

	return envFrom
}

// GetContainersMap gets the map of container name to containers
func GetContainersMap(containers []corev1.Container) map[string]corev1.Container {
	containersMap := make(map[string]corev1.Container)

	for _, container := range containers {
		containersMap[container.Name] = container
	}
	return containersMap
}

// GetPreStartInitContainers gets the init container for every preStart devfile event
func GetPreStartInitContainers(devfile devfileParser.DevfileObj, containers []corev1.Container) []corev1.Container {

	// if there are preStart events, add them as init containers to the podTemplateSpec
	preStartEvents := devfile.Data.GetEvents().PreStart
	var initContainers []corev1.Container
	if len(preStartEvents) > 0 {
		var eventCommands []string
		commandsMap := devfile.Data.GetCommands()
		containersMap := GetContainersMap(containers)

		for _, event := range preStartEvents {
			eventSubCommands := adaptersCommon.GetCommandsFromEvent(commandsMap, strings.ToLower(event))
			eventCommands = append(eventCommands, eventSubCommands...)
		}

		klog.V(4).Infof("PreStart event commands are: %v", strings.Join(eventCommands, ","))

		for i, commandName := range eventCommands {
			if command, ok := commandsMap[commandName]; ok {
				component := command.GetExecComponent()
				commandLine := command.GetExecCommandLine()
				workingDir := command.GetExecWorkingDir()

				var cmdArr []string
				if workingDir != "" {
					// since we are using /bin/sh -c, the command needs to be within a single double quote instance, for example "cd /tmp && pwd"
					cmdArr = []string{adaptersCommon.ShellExecutable, "-c", "cd " + workingDir + " && " + commandLine}
				} else {
					cmdArr = []string{adaptersCommon.ShellExecutable, "-c", commandLine}
				}

				// Get the container info for the given component
				if container, ok := containersMap[component]; ok {
					// override any container command and args with our event command cmdArr
					container.Command = cmdArr
					container.Args = []string{}

					// Override the init container name since there cannot be two containers with the same
					// name in a pod. This applies to pod containers and pod init containers. The convention
					// for init container name here is, containername-eventname-<position of command in prestart events>
					// If there are two events referencing the same devfile component, then we will have
					// tools-event1-1 & tools-event2-3, for example. And if in the edge case, the same command is
					// executed twice by preStart events, then we will have tools-event1-1 & tools-event1-2
					initContainerName := fmt.Sprintf("%s-%s", container.Name, commandName)
					initContainerName = util.TruncateString(initContainerName, containerNameMaxLen)
					initContainerName = fmt.Sprintf("%s-%d", initContainerName, i+1)
					container.Name = initContainerName

					initContainers = append(initContainers, container)
				}
			}
		}

		if len(eventCommands) > 0 {
			log.Successf("PreStart commands have been added to the component: %s", strings.Join(eventCommands, ","))
		}
	}

	return initContainers
}
