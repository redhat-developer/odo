package generator

import (

	// api resource types

	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/apimachinery/pkg/api/resource"

	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
)

const (
	// DevfileSourceVolume is the constant containing the name of the emptyDir volume containing the project source
	DevfileSourceVolume = "devfile-projects"

	// DevfileSourceVolumeMount is the directory to mount the volume in the container
	DevfileSourceVolumeMount = "/projects"

	// EnvProjectsRoot is the env defined for project mount in a component container when component's mountSources=true
	EnvProjectsRoot = "PROJECTS_ROOT"

	// EnvProjectsSrc is the env defined for path to the project source in a component container
	EnvProjectsSrc = "PROJECT_SOURCE"

	deploymentKind       = "Deployment"
	deploymentAPIVersion = "apps/v1"
)

// CreateObjectMeta creates a common object meta
func CreateObjectMeta(name, namespace string, labels, annotations map[string]string) metav1.ObjectMeta {

	objectMeta := metav1.ObjectMeta{
		Name:        name,
		Namespace:   namespace,
		Labels:      labels,
		Annotations: annotations,
	}

	return objectMeta
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

// GenerateContainer creates a container spec that can be used when creating a pod
func GenerateContainer(containerParams ContainerParams) *corev1.Container {
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

// GetContainers iterates through the components in the devfile and returns a slice of the corresponding containers
func GetContainers(devfileObj devfileParser.DevfileObj) ([]corev1.Container, error) {
	var containers []corev1.Container
	for _, comp := range devfileObj.Data.GetComponents() {
		if comp.Container != nil {
			envVars := convertEnvs(comp.Container.Env)
			resourceReqs := getResourceReqs(comp)
			ports, err := convertPorts(comp.Container.Endpoints)
			if err != nil {
				return nil, err
			}
			containerParams := ContainerParams{
				Name:         comp.Name,
				Image:        comp.Container.Image,
				IsPrivileged: false,
				Command:      comp.Container.Command,
				Args:         comp.Container.Args,
				EnvVars:      envVars,
				ResourceReqs: resourceReqs,
				Ports:        ports,
			}
			container := GenerateContainer(containerParams)
			for _, c := range containers {
				for _, containerPort := range c.Ports {
					for _, curPort := range container.Ports {
						if curPort.Name == containerPort.Name {
							// the name has to be unique across containers since it is considered as the URL name
							return nil, fmt.Errorf("devfile contains multiple endpoint entries with same name: %v", containerPort.Name)
						}
						if curPort.ContainerPort == containerPort.ContainerPort {
							// the same TargetPort present in different containers
							// because containers in a single pod shares the network namespace
							return nil, fmt.Errorf("devfile contains multiple containers with same TargetPort: %v", containerPort.ContainerPort)
						}
					}
				}
			}

			// If `mountSources: true` was set, add an empty dir volume to the container to sync the source to
			// Sync to `Container.SourceMapping` if set
			if comp.Container.MountSources {
				var syncRootFolder string
				if comp.Container.SourceMapping != "" {
					syncRootFolder = comp.Container.SourceMapping
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

				// Add PROJECT_SOURCE to the container
				syncFolder, err := GetSyncFolder(syncRootFolder, devfileObj.Data.GetProjects())
				if err != nil {
					return nil, err
				}
				container.Env = append(container.Env,
					corev1.EnvVar{
						Name:  EnvProjectsSrc,
						Value: syncFolder,
					})
			}
			containers = append(containers, *container)
		}
	}
	return containers, nil
}

// PodTemplateSpecParams is a struct that contains the required data to create a pod template spec object
type PodTemplateSpecParams struct {
	ObjectMeta     metav1.ObjectMeta
	InitContainers []corev1.Container
	Containers     []corev1.Container
	Volumes        []corev1.Volume
}

// GeneratePodTemplateSpec creates a pod template spec that can be used to create a deployment spec
func GeneratePodTemplateSpec(podTemplateSpecParams PodTemplateSpecParams) *corev1.PodTemplateSpec {
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

// DeploymentSpecParams is a struct that contains the required data to create a deployment spec object
type DeploymentSpecParams struct {
	PodTemplateSpec   corev1.PodTemplateSpec
	PodSelectorLabels map[string]string
	// ReplicaSet        int32
}

// GenerateDeploymentSpec creates a deployment spec
func GenerateDeploymentSpec(deployParams DeploymentSpecParams) *appsv1.DeploymentSpec {
	// replicaSet := int32(2)
	deploymentSpec := &appsv1.DeploymentSpec{
		Strategy: appsv1.DeploymentStrategy{
			Type: appsv1.RecreateDeploymentStrategyType,
		},
		Selector: &metav1.LabelSelector{
			MatchLabels: deployParams.PodSelectorLabels,
		},
		Template: deployParams.PodTemplateSpec,
		// Replicas: &deployParams.ReplicaSet,
	}

	return deploymentSpec
}

// GeneratePVCSpec creates a pvc spec
func GeneratePVCSpec(quantity resource.Quantity) *corev1.PersistentVolumeClaimSpec {

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

// ServiceSpecParams is a struct that contains the required data to create a svc spec object
type ServiceSpecParams struct {
	SelectorLabels map[string]string
	ContainerPorts []corev1.ContainerPort
}

// GenerateServiceSpec creates a service spec
func GenerateServiceSpec(serviceSpecParams ServiceSpecParams) *corev1.ServiceSpec {
	// generate Service Spec
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

// IngressParams struct for function createIngress
// serviceName is the name of the service for the target reference
// ingressDomain is the ingress domain to use for the ingress
// portNumber is the target port of the ingress
// TLSSecretName is the target TLS Secret name of the ingress
type IngressParams struct {
	ServiceName   string
	IngressDomain string
	PortNumber    intstr.IntOrString
	TLSSecretName string
	Path          string
}

// GenerateIngressSpec creates an ingress spec
func GenerateIngressSpec(ingressParams IngressParams) *extensionsv1.IngressSpec {
	path := "/"
	if ingressParams.Path != "" {
		path = ingressParams.Path
	}
	ingressSpec := &extensionsv1.IngressSpec{
		Rules: []extensionsv1.IngressRule{
			{
				Host: ingressParams.IngressDomain,
				IngressRuleValue: extensionsv1.IngressRuleValue{
					HTTP: &extensionsv1.HTTPIngressRuleValue{
						Paths: []extensionsv1.HTTPIngressPath{
							{
								Path: path,
								Backend: extensionsv1.IngressBackend{
									ServiceName: ingressParams.ServiceName,
									ServicePort: ingressParams.PortNumber,
								},
							},
						},
					},
				},
			},
		},
	}
	secretNameLength := len(ingressParams.TLSSecretName)
	if secretNameLength != 0 {
		ingressSpec.TLS = []extensionsv1.IngressTLS{
			{
				Hosts: []string{
					ingressParams.IngressDomain,
				},
				SecretName: ingressParams.TLSSecretName,
			},
		}
	}

	return ingressSpec
}

// GenerateOwnerReference genertes an ownerReference  from the deployment which can then be set as
// owner for various Kubernetes objects and ensure that when the owner object is deleted from the
// cluster, all other objects are automatically removed by Kubernetes garbage collector
func GenerateOwnerReference(deployment *appsv1.Deployment) metav1.OwnerReference {

	ownerReference := metav1.OwnerReference{
		APIVersion: deploymentAPIVersion,
		Kind:       deploymentKind,
		Name:       deployment.Name,
		UID:        deployment.UID,
	}

	return ownerReference
}
