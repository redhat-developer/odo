package generator

import (
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
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

// GetContainers iterates through the devfile components and returns a slice of the corresponding containers
func GetContainers(devfileObj parser.DevfileObj, options common.DevfileOptions) ([]corev1.Container, error) {
	var containers []corev1.Container
	containerComponents, err := devfileObj.Data.GetDevfileContainerComponents(options)
	if err != nil {
		return nil, err
	}
	for _, comp := range containerComponents {
		envVars := convertEnvs(comp.Container.Env)
		resourceReqs := getResourceReqs(comp)
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
