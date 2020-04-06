package kclient

import (
	"io"
	"time"

	"github.com/golang/glog"
	"github.com/openshift/odo/pkg/log"
	"github.com/pkg/errors"

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
	glog.V(4).Infof("Waiting for %s pod", watchOptions.LabelSelector)
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
				glog.V(4).Infof("Status of %s pod is %s", e.Name, e.Status.Phase)
				switch e.Status.Phase {
				case desiredPhase:
					if !hideSpinner {
						s.End(true)
					}
					glog.V(4).Infof("Pod %s is %v", e.Name, desiredPhase)
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
func (c *Client) ExecCMDInContainer(podName string, containerName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
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
