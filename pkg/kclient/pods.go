package kclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/log"

	// api resource types

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// ExecCMDInContainer execute command in the container of a pod, pass an empty string for containerName to execute in the first container of the pod
func (c *Client) ExecCMDInContainer(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
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
	err = exec.Stream(remotecommand.StreamOptions{
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

// ExtractProjectToComponent extracts the project archive(tar) to the target path from the reader stdin
func (c *Client) ExtractProjectToComponent(containerName, podName string, targetPath string, stdin io.Reader) error {
	// cmdArr will run inside container
	cmdArr := []string{"tar", "xf", "-", "-C", targetPath, "--no-same-owner"}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	klog.V(3).Infof("Executing command %s", strings.Join(cmdArr, " "))
	err := c.ExecCMDInContainer(containerName, podName, cmdArr, &stdout, &stderr, stdin, false)
	if err != nil {
		log.Errorf("Command '%s' in container failed.\n", strings.Join(cmdArr, " "))
		log.Errorf("stdout: %s\n", stdout.String())
		log.Errorf("stderr: %s\n", stderr.String())
		log.Errorf("err: %s\n", err.Error())
		return err
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
		return nil, &PodNotFoundError{Selector: selector}
	} else if numPods > 1 {
		return nil, fmt.Errorf("multiple Pods exist for the selector: %v. Only one must be present", selector)
	}

	// check if the pod is in the terminating state
	if pods.Items[0].DeletionTimestamp != nil {
		return nil, &PodNotFoundError{Selector: selector}
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

func (c *Client) GetAllPodsInNamespace() (*corev1.PodList, error) {
	return c.KubeClient.CoreV1().Pods(c.Namespace).List(context.TODO(), metav1.ListOptions{})
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
