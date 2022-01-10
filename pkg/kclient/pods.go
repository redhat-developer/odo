package kclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/log"
	"k8s.io/klog"

	// api resource types

	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// WaitAndGetPod block and waits until pod matching selector is in in Running state
// desiredPhase cannot be PodFailed or PodUnknown
func (c *Client) WaitAndGetPodWithEvents(selector string, desiredPhase corev1.PodPhase, waitMessage string, pushTimeout time.Duration) (*corev1.Pod, error) {

	klog.V(3).Infof("Waiting for %s pod", selector)

	var spinner *log.Status
	defer func() {
		if spinner != nil {
			spinner.End(false)
		}
	}()

	w, err := c.KubeClient.CoreV1().Pods(c.Namespace).Watch(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to watch pod")
	}
	defer w.Stop()

	// Here we are going to start a loop watching for the pod status
	podChannel := make(chan *corev1.Pod)
	watchErrorChannel := make(chan error)
	failedEvents := make(map[string]corev1.Event)
	go func() {
	loop:
		for {
			val, ok := <-w.ResultChan()
			if !ok {
				watchErrorChannel <- errors.New("watch channel was closed")
				break loop
			}
			if e, ok := val.Object.(*corev1.Pod); ok {
				klog.V(3).Infof("Status of %s pod is %s", e.Name, e.Status.Phase)
				for _, cond := range e.Status.Conditions {
					// using this just for debugging message, so ignoring error on purpose
					jsonCond, _ := json.Marshal(cond)
					klog.V(3).Infof("Pod Conditions: %s", string(jsonCond))
				}
				for _, status := range e.Status.ContainerStatuses {
					// using this just for debugging message, so ignoring error on purpose
					jsonStatus, _ := json.Marshal(status)
					klog.V(3).Infof("Container Status: %s", string(jsonStatus))
				}
				switch e.Status.Phase {
				case desiredPhase:
					klog.V(3).Infof("Pod %s is %v", e.Name, desiredPhase)
					podChannel <- e
					break loop
				case corev1.PodFailed, corev1.PodUnknown:
					watchErrorChannel <- errors.Errorf("pod %s status %s", e.Name, e.Status.Phase)
					break loop
				default:
					// we start in a phase different from the desired one, let's wait
					if spinner == nil {
						spinner = log.Spinner(waitMessage)
						// Collect all the events in a separate go routine
						quit := make(chan int)
						go c.CollectEvents(selector, failedEvents, spinner, quit)
						defer close(quit)
					}
				}
			} else {
				watchErrorChannel <- errors.New("unable to convert event object to Pod")
				break loop
			}
		}
		close(podChannel)
		close(watchErrorChannel)
	}()

	select {
	case val := <-podChannel:
		if spinner != nil {
			spinner.End(true)
		}
		return val, nil
	case err := <-watchErrorChannel:
		return nil, err
	case <-time.After(pushTimeout):

		// Create a useful error if there are any failed events
		errorMessage := fmt.Sprintf(`waited %s but couldn't find running pod matching selector: '%s'`, pushTimeout, selector)

		if len(failedEvents) != 0 {

			tableString := getErrorMessageFromEvents(failedEvents)

			errorMessage = fmt.Sprintf(`waited %s but was unable to find a running pod matching selector: '%s'
For more information to help determine the cause of the error, re-run with '-v'.
See below for a list of failed events that occured more than %d times during deployment:
%s`, pushTimeout, selector, failedEventCount, tableString.String())
		}

		return nil, errors.Errorf(errorMessage)
	}
}

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
		return errors.Wrapf(err, "unable to get Kubernetes client config")
	}

	// Connect to url (constructed from req) using SPDY (HTTP/2) protocol which allows bidirectional streams.
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return errors.Wrapf(err, "unable execute command via SPDY")
	}
	// initialize the transport of the standard shell streams
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Tty:    tty,
	})
	if err != nil {
		return errors.Wrapf(err, "error while streaming command")
	}

	return nil
}

// ExtractProjectToComponent extracts the project archive(tar) to the target path from the reader stdin
func (c *Client) ExtractProjectToComponent(containerName, podName string, targetPath string, stdin io.Reader) error {
	// cmdArr will run inside container
	cmdArr := []string{"tar", "xf", "-", "-C", targetPath}
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

// GetOnePod gets a pod using the component and app name
func (c *Client) GetOnePod(componentName, appName string) (*corev1.Pod, error) {
	return c.GetOnePodFromSelector(componentlabels.GetSelector(componentName, appName))
}

// GetPodUsingComponentName gets a pod using the component name
func (c *Client) GetPodUsingComponentName(componentName string) (*corev1.Pod, error) {
	podSelector := fmt.Sprintf("component=%s", componentName)
	return c.GetOnePodFromSelector(podSelector)
}

// GetOnePodFromSelector gets a pod from the selector
func (c *Client) GetOnePodFromSelector(selector string) (*corev1.Pod, error) {
	pods, err := c.KubeClient.CoreV1().Pods(c.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		// Dont wrap error since we want to know if its a forbidden error
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
		tailLines := int64(1)
		podLogOptions = corev1.PodLogOptions{
			Follow:    true,
			Previous:  false,
			TailLines: &tailLines,
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
