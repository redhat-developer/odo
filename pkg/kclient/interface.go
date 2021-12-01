package kclient

import (
	"io"

	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/go-openapi/spec"
	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/unions"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ClientInterface interface {

	// deployment.go
	GetDeploymentByName(name string) (*appsv1.Deployment, error)
	GetOneDeployment(componentName, appName string) (*appsv1.Deployment, error)
	GetOneDeploymentFromSelector(selector string) (*appsv1.Deployment, error)
	GetDeploymentFromSelector(selector string) ([]appsv1.Deployment, error)
	ListDeployments(selector string) (*appsv1.DeploymentList, error)
	WaitForPodDeletion(name string) error
	WaitForDeploymentRollout(deploymentName string) (*appsv1.Deployment, error)
	CreateDeployment(deploy appsv1.Deployment) (*appsv1.Deployment, error)
	UpdateDeployment(deploy appsv1.Deployment) (*appsv1.Deployment, error)
	ApplyDeployment(deploy appsv1.Deployment) (*appsv1.Deployment, error)
	DeleteDeployment(labels map[string]string) error
	CreateDynamicResource(exampleCustomResource unstructured.Unstructured, gvr *meta.RESTMapping) error
	ListDynamicResource(group, version, resource string) (*unstructured.UnstructuredList, error)
	GetDynamicResource(group, version, resource, name string) (*unstructured.Unstructured, error)
	UpdateDynamicResource(group, version, resource, name string, u *unstructured.Unstructured) error
	DeleteDynamicResource(name, group, version, resource string) error
	LinkSecret(secretName, componentName, applicationName string) error
	UnlinkSecret(secretName, componentName, applicationName string) error
	GetDeploymentLabelValues(label string, selector string) ([]string, error)
	GetDeploymentAPIVersion() (metav1.GroupVersionResource, error)
	IsDeploymentExtensionsV1Beta1() (bool, error)

	// events.go
	CollectEvents(selector string, events map[string]corev1.Event, spinner *log.Status, quit <-chan int)

	// ingress.go
	GetOneIngressFromSelector(selector string) (*unions.KubernetesIngress, error)
	CreateIngress(ingress unions.KubernetesIngress) (*unions.KubernetesIngress, error)
	DeleteIngress(name string) error
	ListIngresses(labelSelector string) (*unions.KubernetesIngressList, error)
	GetIngress(name string) (*unions.KubernetesIngress, error)

	// kclient.go
	GetClient() kubernetes.Interface
	GetConfig() clientcmd.ClientConfig
	GetClientConfig() *rest.Config
	GetDynamicClient() dynamic.Interface
	Delete(labels map[string]string, wait bool) error
	WaitForComponentDeletion(selector string) error
	GeneratePortForwardReq(podName string) *rest.Request
	SetDiscoveryInterface(client discovery.DiscoveryInterface)
	IsResourceSupported(apiGroup, apiVersion, resourceName string) (bool, error)
	IsSSASupported() bool

	// namespace.go
	GetCurrentNamespace() string
	SetNamespace(ns string)
	GetNamespaces() ([]string, error)
	GetNamespace(name string) (*corev1.Namespace, error)
	CreateNamespace(name string) (*corev1.Namespace, error)
	DeleteNamespace(name string, wait bool) error
	SetCurrentNamespace(namespace string) error
	WaitForServiceAccountInNamespace(namespace, serviceAccountName string) error

	// operators.go
	IsServiceBindingSupported() (bool, error)
	IsCSVSupported() (bool, error)
	ListClusterServiceVersions() (*olm.ClusterServiceVersionList, error)
	GetClusterServiceVersion(name string) (olm.ClusterServiceVersion, error)
	GetCustomResourcesFromCSV(csv *olm.ClusterServiceVersion) *[]olm.CRDDescription
	CheckCustomResourceInCSV(customResource string, csv *olm.ClusterServiceVersion) (bool, *olm.CRDDescription)
	SearchClusterServiceVersionList(name string) (*olm.ClusterServiceVersionList, error)
	GetCustomResource(customResource string) (*olm.CRDDescription, error)
	GetCSVWithCR(name string) (*olm.ClusterServiceVersion, error)
	GetResourceSpecDefinition(group, version, kind string) (*spec.Schema, error)
	GetCRDSpec(cr *olm.CRDDescription, resourceType string, resourceName string) (*spec.Schema, error)
	GetRestMappingFromUnstructured(unstructured.Unstructured) (*meta.RESTMapping, error)
	GetOperatorGVRList() ([]meta.RESTMapping, error)

	// pods.go
	WaitAndGetPodWithEvents(selector string, desiredPhase corev1.PodPhase, waitMessage string) (*corev1.Pod, error)
	ExecCMDInContainer(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error
	ExtractProjectToComponent(containerName, podName string, targetPath string, stdin io.Reader) error
	GetOnePod(componentName, appName string) (*corev1.Pod, error)
	GetPodUsingComponentName(componentName string) (*corev1.Pod, error)
	GetOnePodFromSelector(selector string) (*corev1.Pod, error)
	GetPodLogs(podName, containerName string, followLog bool) (io.ReadCloser, error)

	// secrets.go
	CreateTLSSecret(tlsCertificate []byte, tlsPrivKey []byte, objectMeta metav1.ObjectMeta) (*corev1.Secret, error)
	GetSecret(name, namespace string) (*corev1.Secret, error)
	UpdateSecret(secret *corev1.Secret, namespace string) (*corev1.Secret, error)
	DeleteSecret(secretName, namespace string) error
	CreateSecret(objectMeta metav1.ObjectMeta, data map[string]string, ownerReference metav1.OwnerReference) error
	CreateSecrets(componentName string, commonObjectMeta metav1.ObjectMeta, svc *corev1.Service, ownerReference metav1.OwnerReference) error
	ListSecrets(labelSelector string) ([]corev1.Secret, error)
	WaitAndGetSecret(name string, namespace string) (*corev1.Secret, error)

	// service.go
	GetService(name string) (*corev1.Service, error)
	CreateService(svc corev1.Service) (*corev1.Service, error)
	UpdateService(svc corev1.Service) (*corev1.Service, error)
	ListServices(selector string) ([]corev1.Service, error)
	DeleteService(serviceName string) error
	GetOneService(componentName, appName string) (*corev1.Service, error)
	GetOneServiceFromSelector(selector string) (*corev1.Service, error)

	// volumes.go
	CreatePVC(pvc corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error)
	DeletePVC(pvcName string) error
	ListPVCs(selector string) ([]corev1.PersistentVolumeClaim, error)
	ListPVCNames(selector string) ([]string, error)
	GetPVCFromName(pvcName string) (*corev1.PersistentVolumeClaim, error)
	UpdatePVCLabels(pvc *corev1.PersistentVolumeClaim, labels map[string]string) error
	GetAndUpdateStorageOwnerReference(pvc *corev1.PersistentVolumeClaim, ownerReference ...metav1.OwnerReference) error
	UpdateStorageOwnerReference(pvc *corev1.PersistentVolumeClaim, ownerReference ...metav1.OwnerReference) error
}
