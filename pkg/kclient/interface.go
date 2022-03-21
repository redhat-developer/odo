package kclient

import (
	"io"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/go-openapi/spec"
	projectv1 "github.com/openshift/api/project/v1"
	routev1 "github.com/openshift/api/route/v1"
	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/redhat-developer/odo/pkg/kclient/unions"
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

	// all.go

	// GetAllResourcesFromSelector returns all resources of any kind (including CRs) matching the given label selector
	GetAllResourcesFromSelector(selector string, ns string) ([]unstructured.Unstructured, error)

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
	CollectEvents(selector string, events map[string]corev1.Event, quit <-chan int)

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
	GetNamespaceNormal(name string) (*corev1.Namespace, error)
	CreateNamespace(name string) (*corev1.Namespace, error)
	DeleteNamespace(name string, wait bool) error
	SetCurrentNamespace(namespace string) error
	WaitForServiceAccountInNamespace(namespace, serviceAccountName string) error

	// oc_server.go
	GetServerVersion(timeout time.Duration) (*ServerInfo, error)

	// operators.go
	IsServiceBindingSupported() (bool, error)
	IsCSVSupported() (bool, error)
	ListClusterServiceVersions() (*olm.ClusterServiceVersionList, error)
	GetCustomResourcesFromCSV(csv *olm.ClusterServiceVersion) *[]olm.CRDDescription
	GetCSVWithCR(name string) (*olm.ClusterServiceVersion, error)
	GetResourceSpecDefinition(group, version, kind string) (*spec.Schema, error)
	GetRestMappingFromUnstructured(unstructured.Unstructured) (*meta.RESTMapping, error)
	GetOperatorGVRList() ([]meta.RESTMapping, error)

	// pods.go
	WaitAndGetPodWithEvents(selector string, desiredPhase corev1.PodPhase, pushTimeout time.Duration) (*corev1.Pod, error)
	ExecCMDInContainer(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error
	ExtractProjectToComponent(containerName, podName string, targetPath string, stdin io.Reader) error
	GetPodUsingComponentName(componentName string) (*corev1.Pod, error)
	GetOnePodFromSelector(selector string) (*corev1.Pod, error)
	GetPodLogs(podName, containerName string, followLog bool) (io.ReadCloser, error)

	// port_forwarding.go
	// SetupPortForwarding creates port-forwarding for the pod on the port pairs provided in the
	// ["<localhost-port>":"<remote-pod-port>"] format. errOut is used by the client-go library to output any errors
	// encountered while the port-forwarding is running
	SetupPortForwarding(pod *corev1.Pod, portPairs []string, errOut io.Writer) error

	// projects.go
	CreateNewProject(projectName string, wait bool) error
	DeleteProject(name string, wait bool) error
	GetCurrentProjectName() string
	GetProject(projectName string) (*projectv1.Project, error)
	IsProjectSupported() (bool, error)
	ListProjectNames() ([]string, error)

	// routes.go
	IsRouteSupported() (bool, error)
	GetRoute(name string) (*routev1.Route, error)
	CreateRoute(name string, serviceName string, portNumber intstr.IntOrString, labels map[string]string, secureURL bool, path string, ownerReference metav1.OwnerReference) (*routev1.Route, error)
	DeleteRoute(name string) error
	ListRoutes(labelSelector string) ([]routev1.Route, error)
	GetOneRouteFromSelector(selector string) (*routev1.Route, error)

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
	CreateService(svc corev1.Service) (*corev1.Service, error)
	UpdateService(svc corev1.Service) (*corev1.Service, error)
	ListServices(selector string) ([]corev1.Service, error)
	DeleteService(serviceName string) error
	GetOneService(componentName, appName string) (*corev1.Service, error)
	GetOneServiceFromSelector(selector string) (*corev1.Service, error)

	// user.go
	RunLogout(stdout io.Writer) error

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
