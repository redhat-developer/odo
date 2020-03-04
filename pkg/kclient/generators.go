package kclient

import (

	// api resource types
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/apimachinery/pkg/api/resource"
)

// GenerateContainer creates a container spec that can be used when creating a pod
func GenerateContainer(name, image string, isPrivileged bool, command, args []string, envVars []corev1.EnvVar, resourceReqs corev1.ResourceRequirements) *corev1.Container {
	container := &corev1.Container{
		Name:            name,
		Image:           image,
		ImagePullPolicy: corev1.PullAlways,
		Resources:       resourceReqs,
		Command:         command,
		Args:            args,
		Env:             envVars,
	}

	if isPrivileged {
		container.SecurityContext = &corev1.SecurityContext{
			Privileged: &isPrivileged,
		}
	}

	return container
}

// GeneratePodTemplateSpec creates a pod template spec that can be used to create a deployment spec
func GeneratePodTemplateSpec(objectMeta metav1.ObjectMeta, containers []corev1.Container) *corev1.PodTemplateSpec {
	podTemplateSpec := &corev1.PodTemplateSpec{
		ObjectMeta: objectMeta,
		Spec: corev1.PodSpec{
			Containers: containers,
		},
	}

	return podTemplateSpec
}

// GenerateDeploymentSpec creates a deployment spec
func GenerateDeploymentSpec(podTemplateSpec corev1.PodTemplateSpec) *appsv1.DeploymentSpec {
	labels := podTemplateSpec.ObjectMeta.Labels
	deploymentSpec := &appsv1.DeploymentSpec{
		Strategy: appsv1.DeploymentStrategy{
			Type: appsv1.RecreateDeploymentStrategyType,
		},
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		Template: podTemplateSpec,
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

// IngressParameter struct for function createIngress
// serviceName is the name of the service for the target reference
// ingressDomain is the ingress domain to use for the ingress
// portNumber is the target port of the ingress
// TLSSecretName is the target TLS Secret name of the ingress
type IngressParameter struct {
	ServiceName   string
	IngressDomain string
	PortNumber    intstr.IntOrString
	TLSSecretName string
}

// GenerateIngressSpec creates an ingress spec
func GenerateIngressSpec(ingressParam IngressParameter) *extensionsv1.IngressSpec {
	ingressSpec := &extensionsv1.IngressSpec{
		Rules: []extensionsv1.IngressRule{
			{
				Host: ingressParam.IngressDomain,
				IngressRuleValue: extensionsv1.IngressRuleValue{
					HTTP: &extensionsv1.HTTPIngressRuleValue{
						Paths: []extensionsv1.HTTPIngressPath{
							{
								Path: "/",
								Backend: extensionsv1.IngressBackend{
									ServiceName: ingressParam.ServiceName,
									ServicePort: ingressParam.PortNumber,
								},
							},
						},
					},
				},
			},
		},
	}
	secretNameLength := len(ingressParam.TLSSecretName)
	if secretNameLength != 0 {
		ingressSpec.TLS = []extensionsv1.IngressTLS{
			{
				Hosts: []string{
					ingressParam.IngressDomain,
				},
				SecretName: ingressParam.TLSSecretName,
			},
		}
	}

	return ingressSpec
}
