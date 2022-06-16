package logs

import (
	"io"

	"k8s.io/apimachinery/pkg/runtime"

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

func (o *LogsClient) GetLogsForMode(mode string, componentName string, namespace string) ([]map[string]io.ReadCloser, error) {
	var selector string
	logs := []map[string]io.ReadCloser{}
	unknownMode := true

	if mode == odolabels.ComponentDevMode || mode == odolabels.ComponentAnyMode {
		unknownMode = false
		selector = odolabels.GetSelector(componentName, "app", odolabels.ComponentDevMode)
		l, err := o.getLogsWithSelector(selector, namespace)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l...)
	}
	if mode == odolabels.ComponentDeployMode || mode == odolabels.ComponentAnyMode {
		unknownMode = false
		selector = odolabels.GetSelector(componentName, "app", odolabels.ComponentDeployMode)
		l, err := o.getLogsWithSelector(selector, namespace)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l...)
	}
	if unknownMode {
		return nil, InvalidModeError{mode: mode}
	}

	return logs, nil
}

// getLogsWithSelector returns logs for the containers created for resources matching selector in the namespace.
// ignorePods boolean helps get logs for the containers of the independent Pods created in Deploy mode, since they
// don't have an owner unlike the independent Pods created in Dev mode which are owned by the main Deployment created
// by odo dev
func (o *LogsClient) getLogsWithSelector(selector string, namespace string) ([]map[string]io.ReadCloser, error) {
	resources, err := o.kubernetesClient.GetAllResourcesFromSelector(selector, namespace)
	if err != nil {
		return nil, err
	}
	// a set of unique Pods with Pod name as key and the Pod as value; these are the Pods whose logs we get from cluster
	pods := map[string]corev1.Pod{}

	// if there's a Pod in the resources, we add it to the set of Pods whose logs we are interested in
	for _, r := range resources {
		if r.GetKind() == "Pod" {
			var pod corev1.Pod
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(r.Object, &pod)
			if err != nil {
				return nil, err
			}
			pods[pod.GetName()] = pod
		}
	}

	// get all pods in the namespace
	podList, err := o.kubernetesClient.GetAllPodsInNamespace()
	if err != nil {
		return nil, err
	}

	// match pod ownerReference (if any) with resources matching the selector
	for _, pod := range podList.Items {
		for _, owner := range pod.GetOwnerReferences() {
			match, err := o.matchOwnerReferenceWithResources(owner, resources)
			if err != nil {
				return nil, err
			} else if match {
				pods[pod.GetName()] = pod
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
	logs := []map[string]io.ReadCloser{}
	for pod, containers := range podContainersMap {
		for _, container := range containers {
			containerLogs, err := o.kubernetesClient.GetPodLogs(pod, container.Name, false)
			if err != nil {
				return nil, err
			}
			logs = append(logs, map[string]io.ReadCloser{container.Name: containerLogs})
		}
	}

	return logs, nil
}

// matchOwnerReferenceWithResources recursively checks if the owner reference passed to it matches any of the resources
// This is useful when trying to find if a pod is owned by any of the ReplicaSet or Deployment in the cluster.
func (o *LogsClient) matchOwnerReferenceWithResources(owner metav1.OwnerReference, resources []unstructured.Unstructured) (bool, error) {
	// first, check if ownerReference belongs to any of the resources
	for _, resource := range resources {
		if resource.GetUID() != "" && owner.UID != "" && resource.GetUID() == owner.UID {
			return true, nil
		}
	}
	// second, get the resource indicated by ownerReference and check its ownerReferences field
	restMapping, err := o.kubernetesClient.GetRestMappingFromGVK(schema.FromAPIVersionAndKind(owner.APIVersion, owner.Kind))
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
