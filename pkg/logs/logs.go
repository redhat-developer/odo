package logs

import (
	"io"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
	corev1 "k8s.io/api/core/v1"
)

type LogsClient struct {
	kubernetesClient kclient.ClientInterface
}

func NewLogsClient(kubernetesClient kclient.ClientInterface) *LogsClient {
	return &LogsClient{
		kubernetesClient: kubernetesClient,
	}
}

var _ Client = (*LogsClient)(nil)

func (o *LogsClient) DevModeLogs(componentName string, namespace string) (map[string]io.ReadCloser, error) {
	// get all resources in the namespace which are running in Dev mode
	selector := odolabels.Builder().WithComponentName(componentName).WithMode(odolabels.ComponentDevMode).Selector()
	resources, err := o.kubernetesClient.GetAllResourcesFromSelector(selector, namespace)
	if err != nil {
		return nil, err
	}

	// get all pods in the namespace
	podList, err := o.kubernetesClient.GetAllPodsInNamespace()
	if err != nil {
		return nil, err
	}

	// match pod ownerReference (if any) with resources running in Dev mode
	var pods []corev1.Pod
	for _, pod := range podList.Items {
		for _, owner := range pod.GetOwnerReferences() {
			match, err := o.matchOwnerReferenceWithResources(owner, resources)
			if err != nil {
				return nil, err
			} else if match {
				pods = append(pods, pod)
				break // because we don't need to check other owner references of the pod anymore
			}
		}
	}

	// get all containers from the pods of interest
	podContainersMap := map[string][]corev1.Container{}
	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			if _, ok := podContainersMap[pod.Name]; !ok {
				podContainersMap[pod.Name] = []corev1.Container{container}
			} else {
				podContainersMap[pod.Name] = append(podContainersMap[pod.Name], container)
			}
		}
	}

	// get logs of all containers
	logs := map[string]io.ReadCloser{}
	for pod, containers := range podContainersMap {
		for _, container := range containers {
			containerLogs, err := o.kubernetesClient.GetPodLogs(pod, container.Name, false)
			if err != nil {
				return nil, err
			}
			logs[container.Name] = containerLogs
		}
	}

	return logs, nil
}

// matchOwnerReferenceWithResources recursively checks if the owner reference passed to it matches any of the resources
// This is useful when trying to find if a pod is owned by any of the ReplicaSet or Deployment in the cluster.
func (o *LogsClient) matchOwnerReferenceWithResources(owner metav1.OwnerReference, resources []unstructured.Unstructured) (bool, error) {
	// first, check if ownerReference belongs to any of the resources
	for _, resource := range resources {
		if resource.GetUID() == owner.UID {
			return true, nil
		}
	}
	// second, get the resource indicated by ownerReference and check its ownerReferences field
	group, version := getGroupVersion(owner.APIVersion)
	restMapping, err := o.kubernetesClient.GetRestMappingFromGVK(schema.GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    owner.Kind,
	})
	if err != nil {
		return false, err
	}
	resource, err := o.kubernetesClient.GetDynamicResource(restMapping.Resource, owner.Name)
	if err != nil {
		return false, err
	}
	ownerReferences := resource.GetOwnerReferences()
	// recursively check if ownerReference matches any of the resources' UID
	for _, ownerReference := range ownerReferences {
		return o.matchOwnerReferenceWithResources(ownerReference, resources)
	}
	return false, nil
}

func getGroupVersion(apiVersion string) (string, string) {
	var group, version string
	groupVersion := strings.SplitN(apiVersion, "/", 2)
	if len(groupVersion) == 1 {
		// this could be the case where apiVersion only has version info
		group = ""
		version = groupVersion[0]
	} else {
		group = groupVersion[0]
		version = groupVersion[1]
	}
	return group, version
}
