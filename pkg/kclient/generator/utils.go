package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	devfilev1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/openshift/odo/pkg/util"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// convertEnvs converts environment variables from the devfile structure to kubernetes structure
func convertEnvs(vars []devfilev1.EnvVar) []corev1.EnvVar {
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
func convertPorts(endpoints []devfilev1.Endpoint) []corev1.ContainerPort {
	containerPorts := []corev1.ContainerPort{}
	for _, endpoint := range endpoints {
		name := strings.TrimSpace(util.GetDNS1123Name(strings.ToLower(endpoint.Name)))
		name = util.TruncateString(name, 15)

		containerPorts = append(containerPorts, corev1.ContainerPort{
			Name:          name,
			ContainerPort: int32(endpoint.TargetPort),
		})
	}
	return containerPorts
}

// getResourceReqs creates a kubernetes ResourceRequirements object based on resource requirements set in the devfile
func getResourceReqs(comp devfilev1.Component) corev1.ResourceRequirements {
	reqs := corev1.ResourceRequirements{}
	limits := make(corev1.ResourceList)
	if comp.Container != nil && comp.Container.MemoryLimit != "" {
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
func addSyncFolder(container *corev1.Container, sourceVolumePath string, projects []devfilev1.Project) error {
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
func GetPortExposure(containerComponents []devfilev1.Component) map[int]devfilev1.EndpointExposure {
	portExposureMap := make(map[int]devfilev1.EndpointExposure)
	for _, comp := range containerComponents {
		for _, endpoint := range comp.Container.Endpoints {
			// if exposure=public, no need to check for existence
			if endpoint.Exposure == devfilev1.PublicEndpointExposure || endpoint.Exposure == "" {
				portExposureMap[endpoint.TargetPort] = devfilev1.PublicEndpointExposure
			} else if exposure, exist := portExposureMap[endpoint.TargetPort]; exist {
				// if a container has multiple identical ports with different exposure levels, save the highest level in the map
				if endpoint.Exposure == devfilev1.InternalEndpointExposure && exposure == devfilev1.NoneEndpointExposure {
					portExposureMap[endpoint.TargetPort] = devfilev1.InternalEndpointExposure
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
func GetDevfileContainerComponents(data data.DevfileData) []devfilev1.Component {
	var components []devfilev1.Component
	// Only components with aliases are considered because without an alias commands cannot reference them
	for _, comp := range data.GetComponents() {
		if comp.Container != nil {
			components = append(components, comp)
		}
	}
	return components
}

// GetDevfileVolumeComponents iterates through the components in the devfile and returns a map of devfile volume components
func GetDevfileVolumeComponents(data data.DevfileData) []devfilev1.Component {
	var components []devfilev1.Component
	// Only components with aliases are considered because without an alias commands cannot reference them
	for _, comp := range data.GetComponents() {
		if comp.Volume != nil {
			components = append(components, comp)
		}
	}
	return components
}

// ContainerParams is a struct that contains the required data to create a container object
type ContainerParams struct {
	Name         string
	Image        string
	IsPrivileged bool
	Command      []string
	Args         []string
	EnvVars      []corev1.EnvVar
	ResourceReqs corev1.ResourceRequirements
	Ports        []corev1.ContainerPort
}

// getContainer creates a container spec that can be used when creating a pod
func getContainer(containerParams ContainerParams) *corev1.Container {
	container := &corev1.Container{
		Name:            containerParams.Name,
		Image:           containerParams.Image,
		ImagePullPolicy: corev1.PullAlways,
		Resources:       containerParams.ResourceReqs,
		Env:             containerParams.EnvVars,
		Ports:           containerParams.Ports,
		Command:         containerParams.Command,
		Args:            containerParams.Args,
	}

	if containerParams.IsPrivileged {
		container.SecurityContext = &corev1.SecurityContext{
			Privileged: &containerParams.IsPrivileged,
		}
	}

	return container
}

// ServiceSpecParams is a struct that contains the required data to create a svc spec object
type ServiceSpecParams struct {
	SelectorLabels map[string]string
	ContainerPorts []corev1.ContainerPort
}

// getServiceSpec creates a service spec
func getServiceSpec(serviceSpecParams ServiceSpecParams) *corev1.ServiceSpec {
	var svcPorts []corev1.ServicePort
	for _, containerPort := range serviceSpecParams.ContainerPorts {
		svcPort := corev1.ServicePort{

			Name:       containerPort.Name,
			Port:       containerPort.ContainerPort,
			TargetPort: intstr.FromInt(int(containerPort.ContainerPort)),
		}
		svcPorts = append(svcPorts, svcPort)
	}
	svcSpec := &corev1.ServiceSpec{
		Ports:    svcPorts,
		Selector: serviceSpecParams.SelectorLabels,
	}

	return svcSpec
}
