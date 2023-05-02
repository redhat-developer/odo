package kclient

import (
	"context"
	"io"
	"time"

	"github.com/go-openapi/spec"
	projectv1 "github.com/openshift/api/project/v1"
	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	psaApi "k8s.io/pod-security-admission/api"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/platform"
	bindingApi "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	specApi "github.com/redhat-developer/service-binding-operator/apis/spec/v1alpha3"
)

type ClientInterface interface {
	platform.Client

	// binding.go
	IsServiceBindingSupported() (bool, error)
	GetBindableKinds() (bindingApi.BindableKinds, error)
	GetBindableKindStatusRestMapping(bindableKindStatuses []bindingApi.BindableKindsStatus) ([]*meta.RESTMapping, error)
	GetBindingServiceBinding(name string) (bindingApi.ServiceBinding, error)
	GetSpecServiceBinding(name string) (specApi.ServiceBinding, error)
	ListServiceBindingsFromAllGroups() ([]specApi.ServiceBinding, []bindingApi.ServiceBinding, error)
	NewServiceBindingServiceObject(serviceNs string, unstructuredService unstructured.Unstructured, bindingName string) (bindingApi.Service, error)
	GetWorkloadKinds() ([]string, []schema.GroupVersionKind, error)

	// configmap.go
	ListConfigMaps(labelSelector string) ([]corev1.ConfigMap, error)

	// deployment.go
	GetDeploymentByName(name string) (*appsv1.Deployment, error)
	GetOneDeployment(componentName, appName string, isPartOfComponent bool) (*appsv1.Deployment, error)
	GetOneDeploymentFromSelector(selector string) (*appsv1.Deployment, error)
	GetDeploymentFromSelector(selector string) ([]appsv1.Deployment, error)
	CreateDeployment(deploy appsv1.Deployment) (*appsv1.Deployment, error)
	UpdateDeployment(deploy appsv1.Deployment) (*appsv1.Deployment, error)
	ApplyDeployment(deploy appsv1.Deployment) (*appsv1.Deployment, error)
	GetDeploymentAPIVersion() (schema.GroupVersionKind, error)
	IsDeploymentExtensionsV1Beta1() (bool, error)
	DeploymentWatcher(ctx context.Context, selector string) (watch.Interface, error)

	// dynamic.go
	PatchDynamicResource(exampleCustomResource unstructured.Unstructured) (bool, error)
	ListDynamicResources(namespace string, gvr schema.GroupVersionResource, selector string) (*unstructured.UnstructuredList, error)
	GetDynamicResource(gvr schema.GroupVersionResource, name string) (*unstructured.Unstructured, error)
	UpdateDynamicResource(gvr schema.GroupVersionResource, name string, u *unstructured.Unstructured) error
	DeleteDynamicResource(name string, gvr schema.GroupVersionResource, wait bool) error

	// events.go
	PodWarningEventWatcher(ctx context.Context) (result watch.Interface, isForbidden bool, err error)

	// kclient.go
	GetClient() kubernetes.Interface
	GetConfig() clientcmd.ClientConfig
	GetClientConfig() *rest.Config
	GetDynamicClient() dynamic.Interface
	GeneratePortForwardReq(podName string) *rest.Request
	SetDiscoveryInterface(client discovery.DiscoveryInterface)
	IsResourceSupported(apiGroup, apiVersion, resourceName string) (bool, error)
	IsSSASupported() bool
	Refresh() (newConfig bool, err error)

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
	GetCurrentNamespacePolicy() (psaApi.Policy, error)

	// oc_server.go
	GetServerVersion(timeout time.Duration) (*ServerInfo, error)
	GetOCVersion() (string, error)

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
	GetGVRFromGVK(gvk schema.GroupVersionKind) (schema.GroupVersionResource, error)

	// owner_reference.go
	TryWithBlockOwnerDeletion(ownerReference metav1.OwnerReference, exec func(ownerReference metav1.OwnerReference) error) error

	// pods.go
	GetPodUsingComponentName(componentName string) (*corev1.Pod, error)
	PodWatcher(ctx context.Context, selector string) (watch.Interface, error)
	IsPodNameMatchingSelector(ctx context.Context, podname string, selector string) (bool, error)

	// port_forwarding.go
	// SetupPortForwarding creates port-forwarding for the pod on the port pairs provided in the
	// ["<localhost-port>":"<remote-pod-port>"] format. errOut is used by the client-go library to output any errors
	// encountered while the port-forwarding is running
	SetupPortForwarding(pod *corev1.Pod, portPairs []string, out io.Writer, errOut io.Writer, stopChan chan struct{}, address string) error

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
	GetOneService(componentName, appName string, isPartOfComponent bool) (*corev1.Service, error)
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
	UpdateStorageOwnerReference(pvc *corev1.PersistentVolumeClaim, ownerReference ...metav1.OwnerReference) error

	// ingress_routes.go
	ListIngresses(namespace, selector string) (*v1.IngressList, error)

	ListJobs(selector string) (*batchv1.JobList, error)
	// CreateJob creates a K8s job to execute task
	CreateJob(job batchv1.Job, namespace string) (*batchv1.Job, error)
	// WaitForJobToComplete to wait until a job completes or fails; it starts printing log or error if the job does not complete execution after 1 minute
	WaitForJobToComplete(job *batchv1.Job) (*batchv1.Job, error)
	// GetJobLogs retrieves pod logs of a job
	GetJobLogs(job *batchv1.Job, containerName string) (io.ReadCloser, error)
	DeleteJob(jobName string) error

	// registry.go
	// GetRegistryList returns all Devfile registries from both ClusterDevfileRegistriesList (cluster-scoped)
	// and DevfileRegistriesList (namespace-scoped) resources.
	// Registries from DevfileRegistriesList resources are returned first in the list.
	GetRegistryList() ([]api.Registry, error)
}
