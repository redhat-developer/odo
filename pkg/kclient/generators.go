package kclient

import (
	"k8s.io/client-go/rest"

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

// PodTemplateSpecParams is a struct that contains the required data to create a pod template spec object
type PodTemplateSpecParams struct {
	ObjectMeta metav1.ObjectMeta
	Containers []corev1.Container
	Volumes    []corev1.Volume
}

// GeneratePodTemplateSpec creates a pod template spec that can be used to create a deployment spec
func GeneratePodTemplateSpec(podTemplateSpecParams PodTemplateSpecParams) *corev1.PodTemplateSpec {
	podTemplateSpec := &corev1.PodTemplateSpec{
		ObjectMeta: podTemplateSpecParams.ObjectMeta,
		Spec: corev1.PodSpec{
			Containers: podTemplateSpecParams.Containers,
			Volumes:    podTemplateSpecParams.Volumes,
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

// GenerateOwnerReference generates an ownerReference  from the deployment which can then be set as
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

// GeneratePortForwardReq builds a port forward request
func (c *Client) GeneratePortForwardReq(podName string) *rest.Request {
	return c.KubeClient.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Namespace(c.Namespace).
		Name(podName).
		SubResource("portforward")
}
