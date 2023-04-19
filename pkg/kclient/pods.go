package kclient

import (
	"context"
	"fmt"
	"io"

	// api resource types

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/redhat-developer/odo/pkg/platform"
)

// ExecCMDInContainer execute command in the container of a pod, pass an empty string for containerName to execute in the first container of the pod
func (c *Client) ExecCMDInContainer(ctx context.Context, containerName, podName string, cmd []string, stdout, stderr io.Writer, stdin io.Reader, tty bool) error {
	podExecOptions := corev1.PodExecOptions{
		Command: cmd,
		Stdin:   stdin != nil,
		Stdout:  stdout != nil,
		Stderr:  stderr != nil,
		TTY:     tty,
	}

	// If a container name was passed in, set it in the exec options, otherwise leave it blank
	if containerName != "" {
		podExecOptions.Container = containerName
	}

	req := c.KubeClient.CoreV1().RESTClient().
		Post().
		Namespace(c.Namespace).
		Resource("pods").
		Name(podName).
		SubResource("exec").
		VersionedParams(&podExecOptions, scheme.ParameterCodec)

	config, err := c.KubeConfig.ClientConfig()
	if err != nil {
		return fmt.Errorf("unable to get Kubernetes client config: %w", err)
	}

	// Connect to url (constructed from req) using SPDY (HTTP/2) protocol which allows bidirectional streams.
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("unable execute command via SPDY: %w", err)
	}
	// initialize the transport of the standard shell streams
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Tty:    tty,
	})
	if err != nil {
		return fmt.Errorf("error while streaming command: %w", err)
	}

	return nil
}

// GetPodUsingComponentName gets a pod using the component name
func (c *Client) GetPodUsingComponentName(componentName string) (*corev1.Pod, error) {
	podSelector := fmt.Sprintf("component=%s", componentName)
	return c.GetRunningPodFromSelector(podSelector)
}

// GetRunningPodFromSelector gets a pod from the selector
func (c *Client) GetRunningPodFromSelector(selector string) (*corev1.Pod, error) {
	pods, err := c.KubeClient.CoreV1().Pods(c.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
		FieldSelector: "status.phase=Running",
	})
	if err != nil {
		// Don't wrap error since we want to know if it's a forbidden error
		// if we wrap, we lose the err status reason and callers of this api rely on it
		return nil, err
	}
	numPods := len(pods.Items)
	if numPods == 0 {
		return nil, &platform.PodNotFoundError{Selector: selector}
	} else if numPods > 1 {
		return nil, fmt.Errorf("multiple Pods exist for the selector: %v. Only one must be present", selector)
	}

	// check if the pod is in the terminating state
	if pods.Items[0].DeletionTimestamp != nil {
		return nil, &platform.PodNotFoundError{Selector: selector}
	}

	return &pods.Items[0], nil
}

// GetPodLogs prints the log from pod to stdout
func (c *Client) GetPodLogs(podName, containerName string, followLog bool) (io.ReadCloser, error) {

	// Set standard log options
	podLogOptions := corev1.PodLogOptions{Follow: false, Container: containerName}

	// If the log is being followed, set it to follow / don't wait
	if followLog {
		podLogOptions = corev1.PodLogOptions{
			Follow:    true,
			Previous:  false,
			Container: containerName,
		}
	}

	// RESTClient call to kubernetes
	rd, err := c.KubeClient.CoreV1().RESTClient().Get().
		Namespace(c.Namespace).
		Name(podName).
		Resource("pods").
		SubResource("log").
		VersionedParams(&podLogOptions, scheme.ParameterCodec).
		Stream(context.TODO())

	return rd, err
}

func (c *Client) GetAllPodsInNamespaceMatchingSelector(selector string, ns string) (*corev1.PodList, error) {
	podList, err := c.KubeClient.CoreV1().Pods(c.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	resources, err := c.GetAllResourcesFromSelector(selector, ns)
	if err != nil {
		return nil, err
	}

	var list corev1.PodList
	// match pod ownerReference (if any) with resources matching the selector
	for _, pod := range podList.Items {
		for _, owner := range pod.GetOwnerReferences() {
			var match bool
			match, err = matchOwnerReferenceWithResources(c, owner, resources)
			if err != nil {
				return nil, err
			}
			if match {
				list.Items = append(list.Items, pod)
				break // because we don't need to check other owner references of the pod anymore
			}
		}
	}

	return &list, err
}

// matchOwnerReferenceWithResources recursively checks if the owner reference passed to it matches any of the resources
// This is useful when trying to find if a pod is owned by any of the ReplicaSet or Deployment in the cluster.
func matchOwnerReferenceWithResources(c ClientInterface, owner metav1.OwnerReference, resources []unstructured.Unstructured) (bool, error) {
	// first, check if ownerReference belongs to any of the resources
	for _, resource := range resources {
		if resource.GetUID() != "" && owner.UID != "" && resource.GetUID() == owner.UID {
			return true, nil
		}
	}
	// second, get the resource indicated by ownerReference and check its ownerReferences field
	restMapping, err := c.GetRestMappingFromGVK(schema.FromAPIVersionAndKind(owner.APIVersion, owner.Kind))
	if err != nil {
		return false, err
	}
	resource, err := c.GetDynamicResource(restMapping.Resource, owner.Name)
	if err != nil {
		return false, err
	}
	ownerReferences := resource.GetOwnerReferences()
	// recursively check if ownerReference matches any of the resources' UID
	for _, ownerReference := range ownerReferences {
		return matchOwnerReferenceWithResources(c, ownerReference, resources)
	}
	return false, nil
}

func (c *Client) GetPodsMatchingSelector(selector string) (*corev1.PodList, error) {
	return c.KubeClient.CoreV1().Pods(c.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: selector})
}

func (c *Client) PodWatcher(ctx context.Context, selector string) (watch.Interface, error) {
	ns := c.GetCurrentNamespace()
	return c.GetClient().CoreV1().Pods(ns).
		Watch(ctx, metav1.ListOptions{
			LabelSelector: selector,
		})
}

func (c *Client) IsPodNameMatchingSelector(ctx context.Context, podname string, selector string) (bool, error) {
	ns := c.GetCurrentNamespace()
	list, err := c.GetClient().CoreV1().Pods(ns).List(ctx, metav1.ListOptions{
		FieldSelector: "metadata.name=" + podname,
		LabelSelector: selector,
	})
	if err != nil {
		return false, err
	}
	return len(list.Items) > 0, nil
}
