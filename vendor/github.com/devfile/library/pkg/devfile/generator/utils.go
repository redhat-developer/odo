package generator

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"path/filepath"
	"reflect"
	"strings"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	buildv1 "github.com/openshift/api/build/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// convertEnvs converts environment variables from the devfile structure to kubernetes structure
func convertEnvs(vars []v1.EnvVar) []corev1.EnvVar {
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
func convertPorts(endpoints []v1.Endpoint) []corev1.ContainerPort {
	containerPorts := []corev1.ContainerPort{}
	portMap := make(map[string]bool)
	for _, endpoint := range endpoints {
		var portProtocol corev1.Protocol
		portNumber := int32(endpoint.TargetPort)

		if endpoint.Protocol == v1.UDPEndpointProtocol {
			portProtocol = corev1.ProtocolUDP
		} else {
			portProtocol = corev1.ProtocolTCP
		}
		name := endpoint.Name
		if len(name) > 15 {
			// to be compatible with endpoint longer than 15 chars
			name = fmt.Sprintf("port-%v", portNumber)
		}

		if _, exist := portMap[name]; !exist {
			portMap[name] = true
			containerPorts = append(containerPorts, corev1.ContainerPort{
				Name:          name,
				ContainerPort: portNumber,
				Protocol:      portProtocol,
			})
		}
	}
	return containerPorts
}

// getResourceReqs creates a kubernetes ResourceRequirements object based on resource requirements set in the devfile
func getResourceReqs(comp v1.Component) (corev1.ResourceRequirements, error) {
	reqs := corev1.ResourceRequirements{}
	limits := make(corev1.ResourceList)
	requests := make(corev1.ResourceList)
	var returnedErr error
	if comp.Container != nil {
		if comp.Container.MemoryLimit != "" {
			memoryLimit, err := resource.ParseQuantity(comp.Container.MemoryLimit)
			if err != nil {
				errMsg := fmt.Errorf("error parsing memoryLimit requirement for component %s: %v", comp.Name, err.Error())
				returnedErr = multierror.Append(returnedErr, errMsg)
			} else {
				limits[corev1.ResourceMemory] = memoryLimit
			}
		}
		if comp.Container.CpuLimit != "" {
			cpuLimit, err := resource.ParseQuantity(comp.Container.CpuLimit)
			if err != nil {
				errMsg := fmt.Errorf("error parsing cpuLimit requirement for component %s: %v", comp.Name, err.Error())
				returnedErr = multierror.Append(returnedErr, errMsg)
			} else {
				limits[corev1.ResourceCPU] = cpuLimit
			}
		}
		if comp.Container.MemoryRequest != "" {
			memoryRequest, err := resource.ParseQuantity(comp.Container.MemoryRequest)
			if err != nil {
				errMsg := fmt.Errorf("error parsing memoryRequest requirement for component %s: %v", comp.Name, err.Error())
				returnedErr = multierror.Append(returnedErr, errMsg)
			} else {
				requests[corev1.ResourceMemory] = memoryRequest
			}
		}
		if comp.Container.CpuRequest != "" {
			cpuRequest, err := resource.ParseQuantity(comp.Container.CpuRequest)
			if err != nil {
				errMsg := fmt.Errorf("error parsing cpuRequest requirement for component %s: %v", comp.Name, err.Error())
				returnedErr = multierror.Append(returnedErr, errMsg)
			} else {
				requests[corev1.ResourceCPU] = cpuRequest
			}
		}
		if !reflect.DeepEqual(limits, corev1.ResourceList{}) {
			reqs.Limits = limits
		}
		if !reflect.DeepEqual(requests, corev1.ResourceList{}) {
			reqs.Requests = requests
		}
	}
	return reqs, returnedErr
}

// addSyncRootFolder adds the sync root folder to the container env
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

// addSyncFolder adds the sync folder path to the container env
// sourceVolumePath: mount path of the empty dir volume to sync source code
// projects: list of projects from devfile
func addSyncFolder(container *corev1.Container, sourceVolumePath string, projects []v1.Project) error {
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

// containerParams is a struct that contains the required data to create a container object
type containerParams struct {
	Name         string
	Image        string
	IsPrivileged bool
	Command      []string
	Args         []string
	EnvVars      []corev1.EnvVar
	ResourceReqs corev1.ResourceRequirements
	Ports        []corev1.ContainerPort
}

// getContainer gets a container struct that can be used when creating a pod
func getContainer(containerParams containerParams) *corev1.Container {
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

// podTemplateSpecParams is a struct that contains the required data to create a pod template spec object
type podTemplateSpecParams struct {
	ObjectMeta     metav1.ObjectMeta
	InitContainers []corev1.Container
	Containers     []corev1.Container
	Volumes        []corev1.Volume
}

// getPodTemplateSpec gets a pod template spec that can be used to create a deployment spec
func getPodTemplateSpec(podTemplateSpecParams podTemplateSpecParams) *corev1.PodTemplateSpec {
	podTemplateSpec := &corev1.PodTemplateSpec{
		ObjectMeta: podTemplateSpecParams.ObjectMeta,
		Spec: corev1.PodSpec{
			InitContainers: podTemplateSpecParams.InitContainers,
			Containers:     podTemplateSpecParams.Containers,
			Volumes:        podTemplateSpecParams.Volumes,
		},
	}

	return podTemplateSpec
}

// deploymentSpecParams is a struct that contains the required data to create a deployment spec object
type deploymentSpecParams struct {
	PodTemplateSpec   corev1.PodTemplateSpec
	PodSelectorLabels map[string]string
	Replicas          *int32
}

// getDeploymentSpec gets a deployment spec
func getDeploymentSpec(deploySpecParams deploymentSpecParams) *appsv1.DeploymentSpec {
	deploymentSpec := &appsv1.DeploymentSpec{
		Strategy: appsv1.DeploymentStrategy{
			Type: appsv1.RecreateDeploymentStrategyType,
		},
		Selector: &metav1.LabelSelector{
			MatchLabels: deploySpecParams.PodSelectorLabels,
		},
		Template: deploySpecParams.PodTemplateSpec,
		Replicas: deploySpecParams.Replicas,
	}

	return deploymentSpec
}

// getServiceSpec iterates through the devfile components and returns a ServiceSpec
func getServiceSpec(devfileObj parser.DevfileObj, selectorLabels map[string]string, options common.DevfileOptions) (*corev1.ServiceSpec, error) {

	var containerPorts []corev1.ContainerPort
	portExposureMap, err := getPortExposure(devfileObj, options)
	if err != nil {
		return nil, err
	}
	containers, err := GetContainers(devfileObj, options)
	if err != nil {
		return nil, err
	}
	for _, c := range containers {
		for _, port := range c.Ports {
			portExist := false
			for _, entry := range containerPorts {
				if entry.ContainerPort == port.ContainerPort {
					portExist = true
					break
				}
			}
			// if Exposure == none, should not create a service for that port
			if !portExist && portExposureMap[int(port.ContainerPort)] != v1.NoneEndpointExposure {
				containerPorts = append(containerPorts, port)
			}
		}
	}

	var svcPorts []corev1.ServicePort
	for _, containerPort := range containerPorts {
		svcPort := corev1.ServicePort{

			Name:       containerPort.Name,
			Port:       containerPort.ContainerPort,
			TargetPort: intstr.FromInt(int(containerPort.ContainerPort)),
		}
		svcPorts = append(svcPorts, svcPort)
	}
	svcSpec := &corev1.ServiceSpec{
		Ports:    svcPorts,
		Selector: selectorLabels,
	}

	return svcSpec, nil
}

// getPortExposure iterates through all endpoints and returns the highest exposure level of all TargetPort.
// exposure level: public > internal > none
func getPortExposure(devfileObj parser.DevfileObj, options common.DevfileOptions) (map[int]v1.EndpointExposure, error) {
	portExposureMap := make(map[int]v1.EndpointExposure)
	options.ComponentOptions = common.ComponentOptions{
		ComponentType: v1.ContainerComponentType,
	}
	containerComponents, err := devfileObj.Data.GetComponents(options)
	if err != nil {
		return portExposureMap, err
	}
	for _, comp := range containerComponents {
		for _, endpoint := range comp.Container.Endpoints {
			// if exposure=public, no need to check for existence
			if endpoint.Exposure == v1.PublicEndpointExposure || endpoint.Exposure == "" {
				portExposureMap[endpoint.TargetPort] = v1.PublicEndpointExposure
			} else if exposure, exist := portExposureMap[endpoint.TargetPort]; exist {
				// if a container has multiple identical ports with different exposure levels, save the highest level in the map
				if endpoint.Exposure == v1.InternalEndpointExposure && exposure == v1.NoneEndpointExposure {
					portExposureMap[endpoint.TargetPort] = v1.InternalEndpointExposure
				}
			} else {
				portExposureMap[endpoint.TargetPort] = endpoint.Exposure
			}
		}

	}
	return portExposureMap, nil
}

// IngressSpecParams struct for function GenerateIngressSpec
// serviceName is the name of the service for the target reference
// ingressDomain is the ingress domain to use for the ingress
// portNumber is the target port of the ingress
// Path is the path of the ingress
// TLSSecretName is the target TLS Secret name of the ingress
type IngressSpecParams struct {
	ServiceName   string
	IngressDomain string
	PortNumber    intstr.IntOrString
	TLSSecretName string
	Path          string
}

// getIngressSpec gets an ingress spec
func getIngressSpec(ingressSpecParams IngressSpecParams) *extensionsv1.IngressSpec {
	path := "/"
	if ingressSpecParams.Path != "" {
		path = ingressSpecParams.Path
	}
	ingressSpec := &extensionsv1.IngressSpec{
		Rules: []extensionsv1.IngressRule{
			{
				Host: ingressSpecParams.IngressDomain,
				IngressRuleValue: extensionsv1.IngressRuleValue{
					HTTP: &extensionsv1.HTTPIngressRuleValue{
						Paths: []extensionsv1.HTTPIngressPath{
							{
								Path: path,
								Backend: extensionsv1.IngressBackend{
									ServiceName: ingressSpecParams.ServiceName,
									ServicePort: ingressSpecParams.PortNumber,
								},
							},
						},
					},
				},
			},
		},
	}
	secretNameLength := len(ingressSpecParams.TLSSecretName)
	if secretNameLength != 0 {
		ingressSpec.TLS = []extensionsv1.IngressTLS{
			{
				Hosts: []string{
					ingressSpecParams.IngressDomain,
				},
				SecretName: ingressSpecParams.TLSSecretName,
			},
		}
	}

	return ingressSpec
}

// getNetworkingV1IngressSpec gets a networking v1 ingress spec
func getNetworkingV1IngressSpec(ingressSpecParams IngressSpecParams) *networkingv1.IngressSpec {
	path := "/"
	pathTypeImplementationSpecific := networkingv1.PathTypeImplementationSpecific
	if ingressSpecParams.Path != "" {
		path = ingressSpecParams.Path
	}
	ingressSpec := &networkingv1.IngressSpec{
		Rules: []networkingv1.IngressRule{
			{
				Host: ingressSpecParams.IngressDomain,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{
							{
								Path: path,
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: ingressSpecParams.ServiceName,
										Port: networkingv1.ServiceBackendPort{
											Number: ingressSpecParams.PortNumber.IntVal,
										},
									},
								},
								//Field is required to be set based on attempt to create the ingress
								PathType: &pathTypeImplementationSpecific,
							},
						},
					},
				},
			},
		},
	}
	secretNameLength := len(ingressSpecParams.TLSSecretName)
	if secretNameLength != 0 {
		ingressSpec.TLS = []networkingv1.IngressTLS{
			{
				Hosts: []string{
					ingressSpecParams.IngressDomain,
				},
				SecretName: ingressSpecParams.TLSSecretName,
			},
		}
	}

	return ingressSpec
}

// RouteSpecParams struct for function GenerateRouteSpec
// serviceName is the name of the service for the target reference
// portNumber is the target port of the ingress
// Path is the path of the route
type RouteSpecParams struct {
	ServiceName string
	PortNumber  intstr.IntOrString
	Path        string
	Secure      bool
}

// GetRouteSpec gets a route spec
func getRouteSpec(routeParams RouteSpecParams) *routev1.RouteSpec {
	routePath := "/"
	if routeParams.Path != "" {
		routePath = routeParams.Path
	}
	routeSpec := &routev1.RouteSpec{
		To: routev1.RouteTargetReference{
			Kind: "Service",
			Name: routeParams.ServiceName,
		},
		Port: &routev1.RoutePort{
			TargetPort: routeParams.PortNumber,
		},
		Path: routePath,
	}

	if routeParams.Secure {
		routeSpec.TLS = &routev1.TLSConfig{
			Termination:                   routev1.TLSTerminationEdge,
			InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
		}
	}

	return routeSpec
}

// getPVCSpec gets a RWO pvc spec
func getPVCSpec(quantity resource.Quantity) *corev1.PersistentVolumeClaimSpec {

	pvcSpec := &corev1.PersistentVolumeClaimSpec{
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: quantity,
			},
		},
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		},
	}

	return pvcSpec
}

// BuildConfigSpecParams is a struct to create build config spec
type BuildConfigSpecParams struct {
	ImageStreamTagName string
	GitURL             string
	GitRef             string
	ContextDir         string
	BuildStrategy      buildv1.BuildStrategy
}

// getBuildConfigSpec gets the build config spec and outputs the build to the image stream
func getBuildConfigSpec(buildConfigSpecParams BuildConfigSpecParams) *buildv1.BuildConfigSpec {

	return &buildv1.BuildConfigSpec{
		CommonSpec: buildv1.CommonSpec{
			Output: buildv1.BuildOutput{
				To: &corev1.ObjectReference{
					Kind: "ImageStreamTag",
					Name: buildConfigSpecParams.ImageStreamTagName + ":latest",
				},
			},
			Source: buildv1.BuildSource{
				Git: &buildv1.GitBuildSource{
					URI: buildConfigSpecParams.GitURL,
					Ref: buildConfigSpecParams.GitRef,
				},
				ContextDir: buildConfigSpecParams.ContextDir,
				Type:       buildv1.BuildSourceGit,
			},
			Strategy: buildConfigSpecParams.BuildStrategy,
		},
	}
}

// getPVC gets a pvc type volume with the given volume name and pvc name.
func getPVC(volumeName, pvcName string) corev1.Volume {

	return corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvcName,
			},
		},
	}
}

// getEmptyDirVol gets a volume with emptyDir
func getEmptyDirVol(volumeName string) corev1.Volume {
	return corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
}

// addVolumeMountToContainers adds the Volume Mounts in containerNameToMountPaths to the containers for a given volumeName.
// containerNameToMountPaths is a map of a container name to an array of its Mount Paths.
func addVolumeMountToContainers(containers []corev1.Container, volumeName string, containerNameToMountPaths map[string][]string) {

	for containerName, mountPaths := range containerNameToMountPaths {
		for i := range containers {
			if containers[i].Name == containerName {
				for _, mountPath := range mountPaths {
					containers[i].VolumeMounts = append(containers[i].VolumeMounts, corev1.VolumeMount{
						Name:      volumeName,
						MountPath: mountPath,
					},
					)
				}
			}
		}
	}
}

// getAllContainers iterates through the devfile components and returns all container components
func getAllContainers(devfileObj parser.DevfileObj, options common.DevfileOptions) ([]corev1.Container, error) {
	var containers []corev1.Container

	options.ComponentOptions = common.ComponentOptions{
		ComponentType: v1.ContainerComponentType,
	}
	containerComponents, err := devfileObj.Data.GetComponents(options)
	if err != nil {
		return nil, err
	}
	for _, comp := range containerComponents {
		envVars := convertEnvs(comp.Container.Env)
		resourceReqs, err := getResourceReqs(comp)
		if err != nil {
			return containers, err
		}
		ports := convertPorts(comp.Container.Endpoints)
		containerParams := containerParams{
			Name:         comp.Name,
			Image:        comp.Container.Image,
			IsPrivileged: false,
			Command:      comp.Container.Command,
			Args:         comp.Container.Args,
			EnvVars:      envVars,
			ResourceReqs: resourceReqs,
			Ports:        ports,
		}
		container := getContainer(containerParams)

		// If `mountSources: true` was set PROJECTS_ROOT & PROJECT_SOURCE env
		if comp.Container.MountSources == nil || *comp.Container.MountSources {
			syncRootFolder := addSyncRootFolder(container, comp.Container.SourceMapping)

			projects, err := devfileObj.Data.GetProjects(common.DevfileOptions{})
			if err != nil {
				return nil, err
			}
			err = addSyncFolder(container, syncRootFolder, projects)
			if err != nil {
				return nil, err
			}
		}
		containers = append(containers, *container)
	}
	return containers, nil
}

// getContainerAnnotations iterates through container components and returns all annotations
func getContainerAnnotations(devfileObj parser.DevfileObj, options common.DevfileOptions) (v1.Annotation, error) {
	options.ComponentOptions = common.ComponentOptions{
		ComponentType: v1.ContainerComponentType,
	}
	containerComponents, err := devfileObj.Data.GetComponents(options)
	if err != nil {
		return v1.Annotation{}, err
	}
	var annotations v1.Annotation
	annotations.Service = make(map[string]string)
	annotations.Deployment = make(map[string]string)
	for _, comp := range containerComponents {
		// ToDo: dedicatedPod support: https://github.com/devfile/api/issues/670
		if comp.Container.DedicatedPod != nil && *comp.Container.DedicatedPod {
			continue
		}
		if comp.Container.Annotation != nil {
			mergeMaps(annotations.Service, comp.Container.Annotation.Service)
			mergeMaps(annotations.Deployment, comp.Container.Annotation.Deployment)
		}
	}

	return annotations, nil
}

func mergeMaps(dest map[string]string, src map[string]string) map[string]string {
	if dest == nil {
		dest = make(map[string]string)
	}
	for k, v := range src {
		dest[k] = v
	}
	return dest
}
