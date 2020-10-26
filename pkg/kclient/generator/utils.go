package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/openshift/odo/pkg/devfile/parser/data"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/util"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// convertEnvs converts environment variables from the devfile structure to kubernetes structure
func convertEnvs(vars []common.Env) []corev1.EnvVar {
	kVars := []corev1.EnvVar{}
	for _, env := range vars {
		kVars = append(kVars, corev1.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		})
	}
	return kVars
}

// convertPorts converts endpoint variables from the devfile structure to kubernetes ContainerPort
func convertPorts(endpoints []common.Endpoint) ([]corev1.ContainerPort, error) {
	containerPorts := []corev1.ContainerPort{}
	for _, endpoint := range endpoints {
		name := strings.TrimSpace(util.GetDNS1123Name(strings.ToLower(endpoint.Name)))
		name = util.TruncateString(name, 15)
		for _, c := range containerPorts {
			if c.Name == endpoint.Name {
				// the name has to be unique within a single container since it is considered as the URL name
				return nil, fmt.Errorf("devfile contains multiple endpoint entries with same name: %v", endpoint.Name)
			}
		}
		containerPorts = append(containerPorts, corev1.ContainerPort{
			Name:          name,
			ContainerPort: endpoint.TargetPort,
		})
	}
	return containerPorts, nil
}

// getResourceReqs creates a kubernetes ResourceRequirements object based on resource requirements set in the devfile
func getResourceReqs(comp common.DevfileComponent) corev1.ResourceRequirements {
	reqs := corev1.ResourceRequirements{}
	limits := make(corev1.ResourceList)
	if comp.Container.MemoryLimit != "" {
		memoryLimit, err := resource.ParseQuantity(comp.Container.MemoryLimit)
		if err == nil {
			limits[corev1.ResourceMemory] = memoryLimit
		}
		reqs.Limits = limits
	}
	return reqs
}

// addSyncRootFolder adds the sync root folder to the container env and volume mounts
func addSyncRootFolder(container *corev1.Container, sourceMapping string) string {
	var syncRootFolder string
	if sourceMapping != "" {
		syncRootFolder = sourceMapping
	} else {
		syncRootFolder = DevfileSourceVolumeMount
	}

	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
		Name:      DevfileSourceVolume,
		MountPath: syncRootFolder,
	})

	// Note: PROJECTS_ROOT & PROJECT_SOURCE are validated at the devfile parser level
	// Add PROJECTS_ROOT to the container
	container.Env = append(container.Env,
		corev1.EnvVar{
			Name:  EnvProjectsRoot,
			Value: syncRootFolder,
		})

	return syncRootFolder
}

// addSyncFolder adds the sync folder path for the container
// sourceVolumePath: mount path of the empty dir volume to sync source code
// projects: list of projects from devfile
func addSyncFolder(container *corev1.Container, sourceVolumePath string, projects []common.DevfileProject) error {
	var syncFolder string

	// if there are no projects in the devfile, source would be synced to $PROJECTS_ROOT
	if len(projects) == 0 {
		syncFolder = sourceVolumePath
	} else {
		// if there is one or more projects in the devfile, get the first project and check its clonepath
		project := projects[0]
		// If clonepath does not exist source would be synced to $PROJECTS_ROOT/projectName
		syncFolder = filepath.ToSlash(filepath.Join(sourceVolumePath, project.Name))

		if project.ClonePath != "" {
			if strings.HasPrefix(project.ClonePath, "/") {
				return fmt.Errorf("the clonePath %s in the devfile project %s must be a relative path", project.ClonePath, project.Name)
			}
			if strings.Contains(project.ClonePath, "..") {
				return fmt.Errorf("the clonePath %s in the devfile project %s cannot escape the value defined by $PROJECTS_ROOT. Please avoid using \"..\" in clonePath", project.ClonePath, project.Name)
			}
			// If clonepath exist source would be synced to $PROJECTS_ROOT/clonePath
			syncFolder = filepath.ToSlash(filepath.Join(sourceVolumePath, project.ClonePath))
		}
	}

	container.Env = append(container.Env,
		corev1.EnvVar{
			Name:  EnvProjectsSrc,
			Value: syncFolder,
		})

	return nil
}

// GetPortExposure iterate through all endpoints and returns the highest exposure level of all TargetPort.
// exposure level: public > internal > none
// This function should be under parser pkg
func GetPortExposure(containerComponents []common.DevfileComponent) map[int32]common.ExposureType {
	portExposureMap := make(map[int32]common.ExposureType)
	for _, comp := range containerComponents {
		for _, endpoint := range comp.Container.Endpoints {
			// if exposure=public, no need to check for existence
			if endpoint.Exposure == common.Public || endpoint.Exposure == "" {
				portExposureMap[endpoint.TargetPort] = common.Public
			} else if exposure, exist := portExposureMap[endpoint.TargetPort]; exist {
				// if a container has multiple identical ports with different exposure levels, save the highest level in the map
				if endpoint.Exposure == common.Internal && exposure == common.None {
					portExposureMap[endpoint.TargetPort] = common.Internal
				}
			} else {
				portExposureMap[endpoint.TargetPort] = endpoint.Exposure
			}
		}

	}
	return portExposureMap
}

// GetDevfileContainerComponents iterates through the components in the devfile and returns a list of devfile container components
// This function should be under parser pkg
func GetDevfileContainerComponents(data data.DevfileData) []common.DevfileComponent {
	var components []common.DevfileComponent
	// Only components with aliases are considered because without an alias commands cannot reference them
	for _, comp := range data.GetAliasedComponents() {
		if comp.Container != nil {
			components = append(components, comp)
		}
	}
	return components
}
