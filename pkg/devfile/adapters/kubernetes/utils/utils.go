package utils

import (
	"fmt"
	"strconv"
	"strings"

	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/util"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog"
)

// ComponentExists checks whether a deployment by the given name exists
func ComponentExists(client kclient.Client, name string) bool {
	_, err := client.GetDeploymentByName(name)
	return err == nil
}

// ConvertEnvs converts environment variables from the devfile structure to kubernetes structure
func ConvertEnvs(vars []common.Env) []corev1.EnvVar {
	kVars := []corev1.EnvVar{}
	for _, env := range vars {
		kVars = append(kVars, corev1.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		})
	}
	return kVars
}

// ConvertPorts converts endpoint variables from the devfile structure to kubernetes ContainerPort
func ConvertPorts(endpoints []common.Endpoint) ([]corev1.ContainerPort, error) {
	containerPorts := []corev1.ContainerPort{}
	for _, endpoint := range endpoints {
		name := strings.TrimSpace(util.GetDNS1123Name(strings.ToLower(endpoint.Name)))
		name = util.TruncateString(name, 15)
		for _, c := range containerPorts {
			if c.ContainerPort == endpoint.TargetPort {
				return nil, fmt.Errorf("Devfile contains multiple identical ports: %v", endpoint.TargetPort)
			}
		}
		containerPorts = append(containerPorts, corev1.ContainerPort{
			Name:          name,
			ContainerPort: endpoint.TargetPort,
		})
	}
	return containerPorts, nil
}

// GetContainers iterates through the components in the devfile and returns a slice of the corresponding containers
func GetContainers(devfileObj devfileParser.DevfileObj) ([]corev1.Container, error) {
	var containers []corev1.Container
	for _, comp := range adaptersCommon.GetSupportedComponents(devfileObj.Data) {
		envVars := ConvertEnvs(comp.Container.Env)
		resourceReqs := GetResourceReqs(comp)
		ports, err := ConvertPorts(comp.Container.Endpoints)
		if err != nil {
			return nil, err
		}
		container := kclient.GenerateContainer(comp.Container.Name, comp.Container.Image, false, comp.Container.Command, comp.Container.Args, envVars, resourceReqs, ports)
		for _, c := range containers {
			for _, containerPort := range c.Ports {
				for _, curPort := range container.Ports {
					if curPort.ContainerPort == containerPort.ContainerPort {
						return nil, fmt.Errorf("Devfile contains multiple identical ports: %v", containerPort.ContainerPort)
					}
				}
			}
		}

		// If `mountSources: true` was set, add an empty dir volume to the container to sync the source to
		if comp.Container.MountSources {
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				Name:      kclient.OdoSourceVolume,
				MountPath: kclient.OdoSourceVolumeMount,
			})

			// only add the env if it is not set by the devfile
			if !isEnvPresent(container.Env, adaptersCommon.EnvCheProjectsRoot) {
				container.Env = append(container.Env,
					corev1.EnvVar{
						Name:  adaptersCommon.EnvCheProjectsRoot,
						Value: kclient.OdoSourceVolumeMount,
					})
			}
		}
		containers = append(containers, *container)
	}
	return containers, nil
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
			klog.V(4).Infof("Updating container %v with supervisord volume mounts", container.Name)
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				Name:      adaptersCommon.SupervisordVolumeName,
				MountPath: adaptersCommon.SupervisordMountPath,
			})

			// Update the run container's ENV for work dir and command
			// only if the env var is not set in the devfile
			// This is done, so supervisord can use it in it's program
			if !isEnvPresent(container.Env, adaptersCommon.EnvOdoCommandRun) {
				klog.V(4).Infof("Updating container %v env with run command", container.Name)
				container.Env = append(container.Env,
					corev1.EnvVar{
						Name:  adaptersCommon.EnvOdoCommandRun,
						Value: runCommand.Exec.CommandLine,
					})
			}

			if !isEnvPresent(container.Env, adaptersCommon.EnvOdoCommandRunWorkingDir) && runCommand.Exec.WorkingDir != "" {
				klog.V(4).Infof("Updating container %v env with run command's workdir", container.Name)
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
				klog.V(4).Infof("Updating container %v with supervisord volume mounts", container.Name)
				container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
					Name:      adaptersCommon.SupervisordVolumeName,
					MountPath: adaptersCommon.SupervisordMountPath,
				})
			}

			// Update the debug container's ENV for work dir and command
			// only if the env var is not set in the devfile
			// This is done, so supervisord can use it in it's program
			if !isEnvPresent(container.Env, adaptersCommon.EnvOdoCommandDebug) {
				klog.V(4).Infof("Updating container %v env with debug command", container.Name)
				container.Env = append(container.Env,
					corev1.EnvVar{
						Name:  adaptersCommon.EnvOdoCommandDebug,
						Value: debugCommand.Exec.CommandLine,
					})
			}

			if debugCommand.Exec.WorkingDir != "" && !isEnvPresent(container.Env, adaptersCommon.EnvOdoCommandDebugWorkingDir) {
				klog.V(4).Infof("Updating container %v env with debug command's workdir", container.Name)
				container.Env = append(container.Env,
					corev1.EnvVar{
						Name:  adaptersCommon.EnvOdoCommandDebugWorkingDir,
						Value: debugCommand.Exec.WorkingDir,
					})
			}

			if !isEnvPresent(container.Env, adaptersCommon.EnvDebugPort) {
				klog.V(4).Infof("Updating container %v env with debug command's debugPort", container.Name)
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

// GetResourceReqs creates a kubernetes ResourceRequirements object based on resource requirements set in the devfile
func GetResourceReqs(comp common.DevfileComponent) corev1.ResourceRequirements {
	reqs := corev1.ResourceRequirements{}
	limits := make(corev1.ResourceList)
	if &comp.Container.MemoryLimit != nil {
		memoryLimit, err := resource.ParseQuantity(comp.Container.MemoryLimit)
		if err == nil {
			limits[corev1.ResourceMemory] = memoryLimit
		}
		reqs.Limits = limits
	}
	return reqs
}

// overrideContainerArgs overrides the container's entrypoint with supervisord
func overrideContainerArgs(container *corev1.Container) {
	klog.V(4).Infof("Updating container %v entrypoint with supervisord", container.Name)
	container.Command = append(container.Command, adaptersCommon.SupervisordBinaryPath)
	container.Args = append(container.Args, "-c", adaptersCommon.SupervisordConfFile)
}

func PluraliseKind(gvkKind string) (kind string) {
	// gvkKind is normally a singular noun and we need to have the kind as a plural
	// i.e. Deploment => Deployments
	//      Route => Routes
	//      Ingress => Ingresses
	kind = strings.ToLower(gvkKind + "s")
	if strings.HasSuffix(gvkKind, "s") {
		kind = strings.ToLower(gvkKind + "es")
	}
	return kind
}
