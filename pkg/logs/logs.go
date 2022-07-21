package logs

import (
	"fmt"
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

type ContainerLogs struct {
	Name string
	Logs io.ReadCloser
}

type Events struct {
	// channel to put the container logs on
	Logs chan ContainerLogs
	// channel to put an error on, if any
	Err chan error
	// channel to indicate that logs for all pods have been grabbed; not to be populated if --follow is used
	Done chan struct{}
}

var _ Client = (*LogsClient)(nil)

func NewLogsClient(kubernetesClient kclient.ClientInterface) *LogsClient {
	return &LogsClient{
		kubernetesClient: kubernetesClient,
	}
}

var _ Client = (*LogsClient)(nil)

func (o *LogsClient) GetLogsForMode(
	mode string,
	componentName string,
	namespace string,
	follow bool,
) (Events, error) {
	events := Events{
		Logs: make(chan ContainerLogs),
		Err:  make(chan error),
		Done: make(chan struct{}),
	}

	go o.getLogsForMode(events, mode, componentName, namespace, follow)
	return events, nil
}

func (o *LogsClient) getLogsForMode(
	events Events,
	mode string,
	componentName string,
	namespace string,
	follow bool,
) {
	var selector string
	podChan := make(chan corev1.Pod) // grab the logs of the pod put on this channel
	errChan := make(chan error)
	doneChan := make(chan struct{}) // because populating doneChan directly would cause odo logs to exit prematurely.

	go func() {
		// this go routine gets the logs of the pods put on the podChan
		for {
			select {
			case pod := <-podChan:
				for _, container := range pod.Spec.Containers {
					containerLogs, err := o.kubernetesClient.GetPodLogs(pod.Name, container.Name, follow)
					if err != nil {
						events.Err <- fmt.Errorf("failed to get logs for container %s; error: %v", container.Name, err)
					}
					events.Logs <- ContainerLogs{container.Name, containerLogs}
				}
			case err := <-errChan:
				events.Err <- err
			case <-doneChan:
				events.Done <- struct{}{}
			}
		}
	}()

	if mode == odolabels.ComponentDevMode || mode == odolabels.ComponentAnyMode {
		selector = odolabels.GetSelector(componentName, "app", odolabels.ComponentDevMode)
		err := o.getPodsForSelector(selector, namespace, podChan)
		if err != nil {
			errChan <- err
		}
	}
	if mode == odolabels.ComponentDeployMode || mode == odolabels.ComponentAnyMode {
		selector = odolabels.GetSelector(componentName, "app", odolabels.ComponentDeployMode)
		err := o.getPodsForSelector(selector, namespace, podChan)
		if err != nil {
			errChan <- err
		}
	}

	doneChan <- struct{}{}
}

// getPodsForSelector gets pods for the resources matching selector in the namespace; Pods found by this method will be
// put on podChan so that caller function can fetch its logs
func (o *LogsClient) getPodsForSelector(
	selector string,
	namespace string,
	podChan chan corev1.Pod,
) error {
	resources, err := o.kubernetesClient.GetAllResourcesFromSelector(selector, namespace)
	if err != nil {
		return err
	}
	// set of unique Pods with Pod name as key; these are the Pods whose logs we want to get from the cluster
	pods := map[string]struct{}{}

	// if there's a Pod in the resources, we add it to the set of Pods whose logs we are interested in
	for _, r := range resources {
		if r.GetKind() == "Pod" {
			var pod corev1.Pod
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(r.Object, &pod)
			if err != nil {
				return err
			}
			pods[pod.GetName()] = struct{}{}
			if podChan != nil {
				podChan <- pod
			}
		}
	}

	// get all pods in the namespace
	podList, err := o.kubernetesClient.GetAllPodsInNamespace()
	if err != nil {
		return err
	}

	// match pod ownerReference (if any) with resources matching the selector
	for _, pod := range podList.Items {
		match := false
		for _, owner := range pod.GetOwnerReferences() {
			match, err = o.matchOwnerReferenceWithResources(owner, resources)
			if err != nil {
				return err
			} else if match {
				if _, ok := pods[pod.GetName()]; ok {
					// Pod's logs have already been displayed to user
					continue
				}
				pods[pod.GetName()] = struct{}{}
				podChan <- pod
				break // because we don't need to check other owner references of the pod anymore
			}
		}
	}
	return nil
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
