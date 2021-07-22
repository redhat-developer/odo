package generator

import (
	"fmt"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/devfile/library/pkg/util"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// DevfileSourceVolumeMount is the default directory to mount the volume in the container
	DevfileSourceVolumeMount = "/projects"

	// EnvProjectsRoot is the env defined for project mount in a component container when component's mountSources=true
	EnvProjectsRoot = "PROJECTS_ROOT"

	// EnvProjectsSrc is the env defined for path to the project source in a component container
	EnvProjectsSrc = "PROJECT_SOURCE"

	deploymentKind       = "Deployment"
	deploymentAPIVersion = "apps/v1"

	containerNameMaxLen = 55
)

// GetTypeMeta gets a type meta of the specified kind and version
func GetTypeMeta(kind string, APIVersion string) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       kind,
		APIVersion: APIVersion,
	}
}

// GetObjectMeta gets an object meta with the parameters
func GetObjectMeta(name, namespace string, labels, annotations map[string]string) metav1.ObjectMeta {

	objectMeta := metav1.ObjectMeta{
		Name:        name,
		Namespace:   namespace,
		Labels:      labels,
		Annotations: annotations,
	}

	return objectMeta
}

// GetContainers iterates through all container components, filters out init containers and returns corresponding containers
func GetContainers(devfileObj parser.DevfileObj, options common.DevfileOptions) ([]corev1.Container, error) {
	allContainers, err := getAllContainers(devfileObj, options)
	if err != nil {
		return nil, err
	}

	// filter out containers for preStart and postStop events
	preStartEvents := devfileObj.Data.GetEvents().PreStart
	postStopEvents := devfileObj.Data.GetEvents().PostStop
	if len(preStartEvents) > 0 || len(postStopEvents) > 0 {
		var eventCommands []string
		commands, err := devfileObj.Data.GetCommands(common.DevfileOptions{})
		if err != nil {
			return nil, err
		}

		commandsMap := common.GetCommandsMap(commands)

		for _, event := range preStartEvents {
			eventSubCommands := common.GetCommandsFromEvent(commandsMap, event)
			eventCommands = append(eventCommands, eventSubCommands...)
		}

		for _, event := range postStopEvents {
			eventSubCommands := common.GetCommandsFromEvent(commandsMap, event)
			eventCommands = append(eventCommands, eventSubCommands...)
		}

		for _, commandName := range eventCommands {
			command, _ := commandsMap[commandName]
			component := common.GetApplyComponent(command)

			// Get the container info for the given component
			for i, container := range allContainers {
				if container.Name == component {
					allContainers = append(allContainers[:i], allContainers[i+1:]...)
				}
			}
		}
	}

	return allContainers, nil

}

// GetInitContainers gets the init container for every preStart devfile event
func GetInitContainers(devfileObj parser.DevfileObj) ([]corev1.Container, error) {
	containers, err := getAllContainers(devfileObj, common.DevfileOptions{})
	if err != nil {
		return nil, err
	}
	preStartEvents := devfileObj.Data.GetEvents().PreStart
	var initContainers []corev1.Container
	if len(preStartEvents) > 0 {
		var eventCommands []string
		commands, err := devfileObj.Data.GetCommands(common.DevfileOptions{})
		if err != nil {
			return nil, err
		}

		commandsMap := common.GetCommandsMap(commands)

		for _, event := range preStartEvents {
			eventSubCommands := common.GetCommandsFromEvent(commandsMap, event)
			eventCommands = append(eventCommands, eventSubCommands...)
		}

		for i, commandName := range eventCommands {
			command, _ := commandsMap[commandName]
			component := common.GetApplyComponent(command)

			// Get the container info for the given component
			for _, container := range containers {
				if container.Name == component {
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
	}

	return initContainers, nil
}

// DeploymentParams is a struct that contains the required data to create a deployment object
type DeploymentParams struct {
	TypeMeta          metav1.TypeMeta
	ObjectMeta        metav1.ObjectMeta
	InitContainers    []corev1.Container
	Containers        []corev1.Container
	Volumes           []corev1.Volume
	PodSelectorLabels map[string]string
}

// GetDeployment gets a deployment object
func GetDeployment(deployParams DeploymentParams) *appsv1.Deployment {

	podTemplateSpecParams := podTemplateSpecParams{
		ObjectMeta:     deployParams.ObjectMeta,
		InitContainers: deployParams.InitContainers,
		Containers:     deployParams.Containers,
		Volumes:        deployParams.Volumes,
	}

	deploySpecParams := deploymentSpecParams{
		PodTemplateSpec:   *getPodTemplateSpec(podTemplateSpecParams),
		PodSelectorLabels: deployParams.PodSelectorLabels,
	}

	deployment := &appsv1.Deployment{
		TypeMeta:   deployParams.TypeMeta,
		ObjectMeta: deployParams.ObjectMeta,
		Spec:       *getDeploymentSpec(deploySpecParams),
	}

	return deployment
}

// PVCParams is a struct to create PVC
type PVCParams struct {
	TypeMeta   metav1.TypeMeta
	ObjectMeta metav1.ObjectMeta
	Quantity   resource.Quantity
}

// GetPVC returns a PVC
func GetPVC(pvcParams PVCParams) *corev1.PersistentVolumeClaim {
	pvcSpec := getPVCSpec(pvcParams.Quantity)

	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta:   pvcParams.TypeMeta,
		ObjectMeta: pvcParams.ObjectMeta,
		Spec:       *pvcSpec,
	}

	return pvc
}

// ServiceParams is a struct that contains the required data to create a service object
type ServiceParams struct {
	TypeMeta       metav1.TypeMeta
	ObjectMeta     metav1.ObjectMeta
	SelectorLabels map[string]string
}

// GetService gets the service
func GetService(devfileObj parser.DevfileObj, serviceParams ServiceParams, options common.DevfileOptions) (*corev1.Service, error) {

	serviceSpec, err := getServiceSpec(devfileObj, serviceParams.SelectorLabels, options)
	if err != nil {
		return nil, err
	}

	service := &corev1.Service{
		TypeMeta:   serviceParams.TypeMeta,
		ObjectMeta: serviceParams.ObjectMeta,
		Spec:       *serviceSpec,
	}

	return service, nil
}

// IngressParams is a struct that contains the required data to create an ingress object
type IngressParams struct {
	TypeMeta          metav1.TypeMeta
	ObjectMeta        metav1.ObjectMeta
	IngressSpecParams IngressSpecParams
}

// GetIngress gets an ingress
func GetIngress(ingressParams IngressParams) *extensionsv1.Ingress {

	ingressSpec := getIngressSpec(ingressParams.IngressSpecParams)

	ingress := &extensionsv1.Ingress{
		TypeMeta:   ingressParams.TypeMeta,
		ObjectMeta: ingressParams.ObjectMeta,
		Spec:       *ingressSpec,
	}

	return ingress
}

// GetNetworkingV1Ingress gets a networking v1 ingress
func GetNetworkingV1Ingress(ingressParams IngressParams) *networkingv1.Ingress {
	ingressSpec := getNetworkingV1IngressSpec(ingressParams.IngressSpecParams)

	ingress := &networkingv1.Ingress{
		TypeMeta:   ingressParams.TypeMeta,
		ObjectMeta: ingressParams.ObjectMeta,
		Spec:       *ingressSpec,
	}

	return ingress
}

// RouteParams is a struct that contains the required data to create a route object
type RouteParams struct {
	TypeMeta        metav1.TypeMeta
	ObjectMeta      metav1.ObjectMeta
	RouteSpecParams RouteSpecParams
}

// GetRoute gets a route
func GetRoute(routeParams RouteParams) *routev1.Route {

	routeSpec := getRouteSpec(routeParams.RouteSpecParams)

	route := &routev1.Route{
		TypeMeta:   routeParams.TypeMeta,
		ObjectMeta: routeParams.ObjectMeta,
		Spec:       *routeSpec,
	}

	return route
}

// GetOwnerReference generates an ownerReference  from the deployment which can then be set as
// owner for various Kubernetes objects and ensure that when the owner object is deleted from the
// cluster, all other objects are automatically removed by Kubernetes garbage collector
func GetOwnerReference(deployment *appsv1.Deployment) metav1.OwnerReference {

	ownerReference := metav1.OwnerReference{
		APIVersion: deploymentAPIVersion,
		Kind:       deploymentKind,
		Name:       deployment.Name,
		UID:        deployment.UID,
	}

	return ownerReference
}

// BuildConfigParams is a struct that contains the required data to create a build config object
type BuildConfigParams struct {
	TypeMeta              metav1.TypeMeta
	ObjectMeta            metav1.ObjectMeta
	BuildConfigSpecParams BuildConfigSpecParams
}

// GetBuildConfig gets a build config
func GetBuildConfig(buildConfigParams BuildConfigParams) *buildv1.BuildConfig {

	buildConfigSpec := getBuildConfigSpec(buildConfigParams.BuildConfigSpecParams)

	buildConfig := &buildv1.BuildConfig{
		TypeMeta:   buildConfigParams.TypeMeta,
		ObjectMeta: buildConfigParams.ObjectMeta,
		Spec:       *buildConfigSpec,
	}

	return buildConfig
}

// GetSourceBuildStrategy gets the source build strategy
func GetSourceBuildStrategy(imageName, imageNamespace string) buildv1.BuildStrategy {
	return buildv1.BuildStrategy{
		SourceStrategy: &buildv1.SourceBuildStrategy{
			From: corev1.ObjectReference{
				Kind:      "ImageStreamTag",
				Name:      imageName,
				Namespace: imageNamespace,
			},
		},
	}
}

// GetDockerBuildStrategy gets the docker build strategy
func GetDockerBuildStrategy(dockerfilePath string, env []corev1.EnvVar) buildv1.BuildStrategy {
	return buildv1.BuildStrategy{
		Type: buildv1.DockerBuildStrategyType,
		DockerStrategy: &buildv1.DockerBuildStrategy{
			DockerfilePath: dockerfilePath,
			Env:            env,
		},
	}
}

// ImageStreamParams is a struct that contains the required data to create an image stream object
type ImageStreamParams struct {
	TypeMeta   metav1.TypeMeta
	ObjectMeta metav1.ObjectMeta
}

// GetImageStream is a function to return the image stream
func GetImageStream(imageStreamParams ImageStreamParams) imagev1.ImageStream {
	imageStream := imagev1.ImageStream{
		TypeMeta:   imageStreamParams.TypeMeta,
		ObjectMeta: imageStreamParams.ObjectMeta,
	}
	return imageStream
}

// VolumeInfo is a struct to hold the pvc name and the volume name to create a volume.
type VolumeInfo struct {
	PVCName    string
	VolumeName string
}

// VolumeParams is a struct that contains the required data to create Kubernetes Volumes and mount Volumes in Containers
type VolumeParams struct {
	// Containers is a list of containers that needs to be updated for the volume mounts
	Containers []corev1.Container

	// VolumeNameToVolumeInfo is a map of the devfile volume name to the volume info containing the pvc name and the volume name.
	VolumeNameToVolumeInfo map[string]VolumeInfo
}

// GetVolumesAndVolumeMounts gets the PVC volumes and updates the containers with the volume mounts.
func GetVolumesAndVolumeMounts(devfileObj parser.DevfileObj, volumeParams VolumeParams, options common.DevfileOptions) ([]corev1.Volume, error) {

	options.ComponentOptions = common.ComponentOptions{
		ComponentType: v1.ContainerComponentType,
	}
	containerComponents, err := devfileObj.Data.GetComponents(options)
	if err != nil {
		return nil, err
	}

	var pvcVols []corev1.Volume
	for volName, volInfo := range volumeParams.VolumeNameToVolumeInfo {
		pvcVols = append(pvcVols, getPVC(volInfo.VolumeName, volInfo.PVCName))

		// containerNameToMountPaths is a map of the Devfile container name to their Devfile Volume Mount Paths for a given Volume Name
		containerNameToMountPaths := make(map[string][]string)
		for _, containerComp := range containerComponents {
			for _, volumeMount := range containerComp.Container.VolumeMounts {
				if volName == volumeMount.Name {
					containerNameToMountPaths[containerComp.Name] = append(containerNameToMountPaths[containerComp.Name], GetVolumeMountPath(volumeMount))
				}
			}
		}

		addVolumeMountToContainers(volumeParams.Containers, volInfo.VolumeName, containerNameToMountPaths)
	}
	return pvcVols, nil
}

// GetVolumeMountPath gets the volume mount's path.
func GetVolumeMountPath(volumeMount v1.VolumeMount) string {
	// if there is no volume mount path, default to volume mount name as per devfile schema
	if volumeMount.Path == "" {
		volumeMount.Path = "/" + volumeMount.Name
	}

	return volumeMount.Path
}
