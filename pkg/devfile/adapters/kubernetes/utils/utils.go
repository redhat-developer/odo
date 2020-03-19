package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/openshift/odo/pkg/devfile"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/versions/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/util"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	volumeSize = "5Gi"
)

// ComponentExists checks whether a deployment by the given name exists
func ComponentExists(client kclient.Client, name string) bool {
	_, err := client.GetDeploymentByName(name)
	return err == nil
}

// ConvertEnvs converts environment variables from the devfile structure to kubernetes structure
func ConvertEnvs(vars []common.DockerimageEnv) []corev1.EnvVar {
	kVars := []corev1.EnvVar{}
	for _, env := range vars {
		kVars = append(kVars, corev1.EnvVar{
			Name:  *env.Name,
			Value: *env.Value,
		})
	}
	return kVars
}

// ConvertPorts converts endpoint variables from the devfile structure to kubernetes ContainerPort
func ConvertPorts(endpoints []common.DockerimageEndpoint) ([]corev1.ContainerPort, error) {
	containerPorts := []corev1.ContainerPort{}
	for _, endpoint := range endpoints {
		name := strings.TrimSpace(util.GetDNS1123Name(strings.ToLower(*endpoint.Name)))
		name = util.TruncateString(name, 15)
		for _, c := range containerPorts {
			if c.ContainerPort == *endpoint.Port {
				return nil, fmt.Errorf("Devfile contains multiple identical ports: %v", *endpoint.Port)
			}
		}
		containerPorts = append(containerPorts, corev1.ContainerPort{
			Name:          name,
			ContainerPort: *endpoint.Port,
		})
	}
	return containerPorts, nil
}

// GetContainers iterates through the components in the devfile and returns a slice of the corresponding containers
func GetContainers(devfileObj devfile.DevfileObj) ([]corev1.Container, error) {
	var containers []corev1.Container
	for _, comp := range adaptersCommon.GetSupportedComponents(devfileObj.Data) {
		envVars := ConvertEnvs(comp.Env)
		resourceReqs := GetResourceReqs(comp)
		ports, err := ConvertPorts(comp.Endpoints)
		if err != nil {
			return nil, err
		}
		container := kclient.GenerateContainer(*comp.Alias, *comp.Image, false, comp.Command, comp.Args, envVars, resourceReqs, ports)
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
		if comp.MountSources {
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				Name:      kclient.OdoSourceVolume,
				MountPath: kclient.OdoSourceVolumeMount,
			})
		}
		containers = append(containers, *container)
	}
	return containers, nil
}

// UpdateContainersWithSupervisordIfReqd updates the run components entrypoint and volume mount
// with supervisord if no entrypoint has been specified for the component in the devfile
func UpdateContainersWithSupervisordIfReqd(devfileObj devfile.DevfileObj, containers []corev1.Container, devfileRunCmd string) []corev1.Container {
	// mountSupervisordVolume := false
	runCommandComponents := adaptersCommon.GetRunCommandComponents(devfileObj.Data, devfileRunCmd)
	glog.V(3).Infof("mjf run cmd components %v", runCommandComponents)

	for i, container := range containers {
		for _, runCommandComponent := range runCommandComponents {
			// Check if the container belongs to a run command component
			if reflect.DeepEqual(container.Name, runCommandComponent) {
				// If the run component container has no entrypoint and arguments, override the entrypoint with supervisord
				if len(container.Command) == 0 && len(container.Args) == 0 {
					glog.V(3).Infof("mjf updating container %v", container.Name)
					container.Command = append(container.Command, "/opt/odo/bin/supervisord")
					container.Args = append(container.Args, "-c", "/opt/odo/conf/devfile-supervisor.conf")
				}

				// Always mount the supervisord volume in the run component container
				container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
					Name:      kclient.GetSupervisordVolumeName(),
					MountPath: "/opt/odo/",
				})

				// Update the containers array since the array is not a pointer to the container
				containers[i] = container
			}
		}
	}
	glog.V(3).Infof("mjf container before check %v", containers)
	return containers

}

// GetVolumes iterates through the components in the devfile and returns a map of component alias to the devfile volumes
func GetVolumes(devfileObj devfile.DevfileObj) map[string][]adaptersCommon.DevfileVolume {
	// componentAliasToVolumes is a map of the Devfile Component Alias to the Devfile Component Volumes
	componentAliasToVolumes := make(map[string][]adaptersCommon.DevfileVolume)
	size := volumeSize
	for _, comp := range adaptersCommon.GetSupportedComponents(devfileObj.Data) {
		if comp.Volumes != nil {
			for _, volume := range comp.Volumes {
				vol := adaptersCommon.DevfileVolume{
					Name:          volume.Name,
					ContainerPath: volume.ContainerPath,
					Size:          &size,
				}
				componentAliasToVolumes[*comp.Alias] = append(componentAliasToVolumes[*comp.Alias], vol)
			}
		}
	}
	return componentAliasToVolumes
}

// GetResourceReqs creates a kubernetes ResourceRequirements object based on resource requirements set in the devfile
func GetResourceReqs(comp common.DevfileComponent) corev1.ResourceRequirements {
	reqs := corev1.ResourceRequirements{}
	limits := make(corev1.ResourceList)
	if comp.MemoryLimit != nil {
		memoryLimit, err := resource.ParseQuantity(*comp.MemoryLimit)
		if err == nil {
			limits[corev1.ResourceMemory] = memoryLimit
		}
		reqs.Limits = limits
	}
	return reqs
}
