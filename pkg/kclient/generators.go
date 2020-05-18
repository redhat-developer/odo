package kclient

import (
	"github.com/openshift/odo/pkg/devfile/adapters/common"

	// api resource types

	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	// OdoSourceVolume is the constant containing the name of the emptyDir volume containing the project source
	OdoSourceVolume = "odo-projects"

	// OdoSourceVolumeMount is the directory to mount the volume in the container
	OdoSourceVolumeMount = "/projects"
)

// GenerateContainer creates a container spec that can be used when creating a pod
func GenerateContainer(name, image string, isPrivileged bool, envVars []corev1.EnvVar, resourceReqs corev1.ResourceRequirements, ports []corev1.ContainerPort) *corev1.Container {
	container := &corev1.Container{
		Name:            name,
		Image:           image,
		ImagePullPolicy: corev1.PullAlways,
		Resources:       resourceReqs,
		Env:             envVars,
		Ports:           ports,
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
			Volumes: []corev1.Volume{
				{
					Name: OdoSourceVolume,
				},
				{
					// Create a volume that will be shared betwen InitContainer and the applicationContainer
					// in order to pass over the SupervisorD binary
					Name: common.SupervisordVolumeName,
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
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

// GenerateServiceSpec creates a service spec
func GenerateServiceSpec(componentName string, containerPorts []corev1.ContainerPort) *corev1.ServiceSpec {
	// generate Service Spec
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
		Ports: svcPorts,
		Selector: map[string]string{
			"component": componentName,
		},
	}

	return svcSpec
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

// SelfSignedCertificate struct is the return type of function GenerateSelfSignedCertificate
// CertPem is the byte array for certificate pem encode
// KeyPem is the byte array for key pem encode
type SelfSignedCertificate struct {
	CertPem []byte
	KeyPem  []byte
}

// GenerateSelfSignedCertificate creates a self-signed SSl certificate
func GenerateSelfSignedCertificate(host string) (SelfSignedCertificate, error) {

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return SelfSignedCertificate{}, errors.Wrap(err, "unable to generate rsa key")
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName:   "Odo self-signed certificate",
			Organization: []string{"Odo"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 365 * 10),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		DNSNames:              []string{"*." + host},
	}

	certificateDerEncoding, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return SelfSignedCertificate{}, errors.Wrap(err, "unable to create certificate")
	}
	out := &bytes.Buffer{}
	err = pem.Encode(out, &pem.Block{Type: "CERTIFICATE", Bytes: certificateDerEncoding})
	if err != nil {
		return SelfSignedCertificate{}, errors.Wrap(err, "unable to encode certificate")
	}
	certPemEncode := out.String()
	certPemByteArr := []byte(certPemEncode)

	tlsPrivKeyEncoding := x509.MarshalPKCS1PrivateKey(privateKey)
	err = pem.Encode(out, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: tlsPrivKeyEncoding})
	if err != nil {
		return SelfSignedCertificate{}, errors.Wrap(err, "unable to encode rsa private key")
	}
	keyPemEncode := out.String()
	keyPemByteArr := []byte(keyPemEncode)

	return SelfSignedCertificate{CertPem: certPemByteArr, KeyPem: keyPemByteArr}, nil
}

// GenerateOwnerReference genertes an ownerReference  from the deployment which can then be set as
// owner for various Kubernetes objects and ensure that when the owner object is deleted from the
// cluster, all other objects are automatically removed by Kubernetes garbage collector
func GenerateOwnerReference(deployment *appsv1.Deployment) metav1.OwnerReference {

	ownerReference := metav1.OwnerReference{
		APIVersion: DeploymentAPIVersion,
		Kind:       DeploymentKind,
		Name:       deployment.Name,
		UID:        deployment.UID,
	}

	return ownerReference
}
