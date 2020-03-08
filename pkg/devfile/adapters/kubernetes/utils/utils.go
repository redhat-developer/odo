package utils

import (
	"strings"

	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/devfile/versions/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/util"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
func ConvertPorts(endpoints []common.DockerimageEndpoint) []corev1.ContainerPort {
	containerPorts := []corev1.ContainerPort{}
	for _, endpoint := range endpoints {
		name := strings.TrimSpace(util.GetDNS1123Name(strings.ToLower(*endpoint.Name)))
		containerPorts = append(containerPorts, corev1.ContainerPort{
			Name:          name,
			ContainerPort: *endpoint.Port,
		})
	}
	return containerPorts
}

// GetContainers iterates through the components in the devfile and returns a slice of the corresponding containers
func GetContainers(devfileObj devfile.DevfileObj) []corev1.Container {
	var containers []corev1.Container
	// Only components with aliases are considered because without an alias commands cannot reference them
	for _, comp := range devfileObj.Data.GetAliasedComponents() {
		if comp.Type == common.DevfileComponentTypeDockerimage {
			glog.V(3).Infof("Found component %v with alias %v\n", comp.Type, *comp.Alias)
			envVars := ConvertEnvs(comp.Env)
			resourceReqs := GetResourceReqs(comp)
			ports := ConvertPorts(comp.Endpoints)
			container := kclient.GenerateContainer(*comp.Alias, *comp.Image, false, comp.Command, comp.Args, envVars, resourceReqs, ports)
			containers = append(containers, *container)
		}
	}
	return containers
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
