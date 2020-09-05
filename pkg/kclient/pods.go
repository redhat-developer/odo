package kclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/log"
	"github.com/pkg/errors"
	"k8s.io/klog"

	// api resource types

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	// waitForPodTimeOut controls how long we should wait for a pod before giving up
	waitForPodTimeOut = 240 * time.Second
)

// WaitAndGetPod block and waits until pod matching selector is in the desired phase
// desiredPhase cannot be PodFailed or PodUnknown
func (c *Client) WaitAndGetPod(watchOptions metav1.ListOptions, desiredPhase corev1.PodPhase, waitMessage string, hideSpinner bool) (*corev1.Pod, error) {
	klog.V(3).Infof("Waiting for %s pod", watchOptions.LabelSelector)
	var s *log.Status
	if !hideSpinner {
		s = log.Spinner(waitMessage)
		defer s.End(false)
	}

	w, err := c.KubeClient.CoreV1().Pods(c.Namespace).Watch(watchOptions)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to watch pod")
	}
	defer w.Stop()

	podChannel := make(chan *corev1.Pod)
	watchErrorChannel := make(chan error)
	go func() {
		defer close(podChannel)
		defer close(watchErrorChannel)

		for {
			val, ok := <-w.ResultChan()
			if !ok {
				watchErrorChannel <- errors.New("watch channel was closed")
				return
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
					if !hideSpinner {
						s.End(true)
					}
					klog.V(3).Infof("Pod %s is %v", e.Name, desiredPhase)
					podChannel <- e
					return
				case corev1.PodFailed, corev1.PodUnknown:
					watchErrorChannel <- errors.Errorf("pod %s status %s", e.Name, e.Status.Phase)
					return
				}
			} else {
				watchErrorChannel <- errors.New("unable to convert event object to Pod")
				return
			}
		}
	}()

	select {
	case val := <-podChannel:
		return val, nil
	case err := <-watchErrorChannel:
		return nil, err
	case <-time.After(waitForPodTimeOut):
		return nil, errors.Errorf("waited %s but couldn't find running pod matching selector: '%s'", waitForPodTimeOut, watchOptions.LabelSelector)
	}
}

// ExecCMDInContainer execute command in the container of a pod, pass an empty string for containerName to execute in the first container of the pod
func (c *Client) ExecCMDInContainer(compInfo common.ComponentInfo, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
	podExecOptions := corev1.PodExecOptions{
		Command: cmd,
		Stdin:   stdin != nil,
		Stdout:  stdout != nil,
		Stderr:  stderr != nil,
		TTY:     tty,
	}

	// If a container name was passed in, set it in the exec options, otherwise leave it blank
	if compInfo.ContainerName != "" {
		podExecOptions.Container = compInfo.ContainerName
	}

	req := c.KubeClient.CoreV1().RESTClient().
		Post().
		Namespace(c.Namespace).
		Resource("pods").
		Name(compInfo.PodName).
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
func (c *Client) ExtractProjectToComponent(compInfo common.ComponentInfo, targetPath string, stdin io.Reader) error {
	// cmdArr will run inside container
	cmdArr := []string{"tar", "xf", "-", "-C", targetPath}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	klog.V(3).Infof("Executing command %s", strings.Join(cmdArr, " "))
	err := c.ExecCMDInContainer(compInfo, cmdArr, &stdout, &stderr, stdin, false)
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
	return c.GetOnePodFromSelector(podSelector)
}

// GetOnePodFromSelector gets a pod from the selector
func (c *Client) GetOnePodFromSelector(selector string) (*corev1.Pod, error) {
	pods, err := c.KubeClient.CoreV1().Pods(c.Namespace).List(metav1.ListOptions{
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

	return &pods.Items[0], nil
}

// GetPodLogs prints the log from pod to stdout
func (c *Client) GetPodLogs(podName, containerName string, followLog bool) (io.ReadCloser, error) {

	// Set standard log options
	podLogOptions := corev1.PodLogOptions{Follow: false}

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
		Stream()

	return rd, err
}
