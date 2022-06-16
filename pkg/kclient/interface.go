package kclient

import (
	"io"
	"time"

	"github.com/go-openapi/spec"
	projectv1 "github.com/openshift/api/project/v1"
	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	bindingApi "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	specApi "github.com/redhat-developer/service-binding-operator/apis/spec/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

	// binding.go
	IsServiceBindingSupported() (bool, error)
	GetBindableKinds() (bindingApi.BindableKinds, error)
	GetBindableKindStatusRestMapping(bindableKindStatuses []bindingApi.BindableKindsStatus) ([]*meta.RESTMapping, error)
	GetBindingServiceBinding(name string) (bindingApi.ServiceBinding, error)
	GetSpecServiceBinding(name string) (specApi.ServiceBinding, error)
	ListServiceBindingsFromAllGroups() ([]specApi.ServiceBinding, []bindingApi.ServiceBinding, error)
	NewServiceBindingServiceObject(unstructuredService unstructured.Unstructured, bindingName string) (bindingApi.Service, error)

	// deployment.go
	GetDeploymentByName(name string) (*appsv1.Deployment, error)
	GetOneDeployment(componentName, appName string) (*appsv1.Deployment, error)
	GetOneDeploymentFromSelector(selector string) (*appsv1.Deployment, error)
	GetDeploymentFromSelector(selector string) ([]appsv1.Deployment, error)
	WaitForPodDeletion(name string) error
	WaitForDeploymentRollout(deploymentName string) (*appsv1.Deployment, error)
	CreateDeployment(deploy appsv1.Deployment) (*appsv1.Deployment, error)
	UpdateDeployment(deploy appsv1.Deployment) (*appsv1.Deployment, error)
	ApplyDeployment(deploy appsv1.Deployment) (*appsv1.Deployment, error)
	GetDeploymentAPIVersion() (schema.GroupVersionResource, error)
	IsDeploymentExtensionsV1Beta1() (bool, error)

	// dynamic.go
	PatchDynamicResource(exampleCustomResource unstructured.Unstructured) (bool, error)
	ListDynamicResources(gvr schema.GroupVersionResource) (*unstructured.UnstructuredList, error)
	GetDynamicResource(gvr schema.GroupVersionResource, name string) (*unstructured.Unstructured, error)
	UpdateDynamicResource(gvr schema.GroupVersionResource, name string, u *unstructured.Unstructured) error
	DeleteDynamicResource(name string, gvr schema.GroupVersionResource, wait bool) error

	// events.go
	CollectEvents(selector string, events map[string]corev1.Event, quit <-chan int)

	// kclient.go
	GetClient() kubernetes.Interface
	GetConfig() clientcmd.ClientConfig
	GetClientConfig() *rest.Config
	GetDynamicClient() dynamic.Interface
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
	IsCSVSupported() (bool, error)
	ListClusterServiceVersions() (*olm.ClusterServiceVersionList, error)
	GetCustomResourcesFromCSV(csv *olm.ClusterServiceVersion) *[]olm.CRDDescription
	GetCSVWithCR(name string) (*olm.ClusterServiceVersion, error)
	GetResourceSpecDefinition(group, version, kind string) (*spec.Schema, error)
	GetRestMappingFromUnstructured(unstructured.Unstructured) (*meta.RESTMapping, error)
	GetRestMappingFromGVK(gvk schema.GroupVersionKind) (*meta.RESTMapping, error)
	GetOperatorGVRList() ([]meta.RESTMapping, error)
	GetGVKFromGVR(gvr schema.GroupVersionResource) (schema.GroupVersionKind, error)

	// owner_reference.go
	TryWithBlockOwnerDeletion(ownerReference metav1.OwnerReference, exec func(ownerReference metav1.OwnerReference) error) error

	// pods.go
	WaitAndGetPodWithEvents(selector string, desiredPhase corev1.PodPhase, pushTimeout time.Duration) (*corev1.Pod, error)
	ExecCMDInContainer(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error
	ExtractProjectToComponent(containerName, podName string, targetPath string, stdin io.Reader) error
	GetPodUsingComponentName(componentName string) (*corev1.Pod, error)
	GetOnePodFromSelector(selector string) (*corev1.Pod, error)
	GetPodLogs(podName, containerName string, followLog bool) (io.ReadCloser, error)
	GetAllPodsInNamespace() (*corev1.PodList, error)

	// port_forwarding.go
	// SetupPortForwarding creates port-forwarding for the pod on the port pairs provided in the
	// ["<localhost-port>":"<remote-pod-port>"] format. errOut is used by the client-go library to output any errors
	// encountered while the port-forwarding is running
	SetupPortForwarding(pod *corev1.Pod, portPairs []string, out io.Writer, errOut io.Writer) error

	// projects.go
	CreateNewProject(projectName string, wait bool) error
	DeleteProject(name string, wait bool) error
	GetCurrentProjectName() string
	GetProject(projectName string) (*projectv1.Project, error)
	IsProjectSupported() (bool, error)
	ListProjectNames() ([]string, error)

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
