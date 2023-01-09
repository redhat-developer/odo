package podman

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/redhat-developer/odo/pkg/platform"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog"
)

// GetPodsMatchingSelector returns all pods matching the given label selector.
func (o *PodmanCli) GetPodsMatchingSelector(selector string) (*corev1.PodList, error) {
	// TODO(feloy) when pod is created with labels
	return nil, nil
}

// GetAllResourcesFromSelector returns all resources of any kind matching the given label selector.
func (o *PodmanCli) GetAllResourcesFromSelector(selector string, _ string) ([]unstructured.Unstructured, error) {
	list, err := o.getPodsFromSelector(selector)
	if err != nil {
		return nil, err
	}
	for _, pod := range list {
		klog.V(5).Infof("\npod name: %s", pod.Name)
		klog.V(5).Infof("labels:")
		for k, v := range pod.Labels {
			klog.V(5).Infof(" - %s: %s", k, v)
		}
	}

	var result []unstructured.Unstructured
	for _, pod := range list {
		u := unstructured.Unstructured{}
		u.SetName(pod.Name)
		u.SetLabels(pod.Labels)
		result = append(result, u)
	}

	return result, nil
}

// GetAllPodsInNamespaceMatchingSelector returns all pods matching the given label selector and in the specified namespace.
func (o *PodmanCli) GetAllPodsInNamespaceMatchingSelector(selector string, ns string) (*corev1.PodList, error) {
	// TODO(feloy) when pod is created with labels
	return nil, nil
}

// GetRunningPodFromSelector returns any pod matching the given label selector.
// If multiple pods are found, implementations might have different behavior, by either returning an error or returning any element.
func (o *PodmanCli) GetRunningPodFromSelector(selector string) (*corev1.Pod, error) {
	list, err := o.getPodsFromSelector(selector)
	if err != nil {
		return nil, err
	}
	numPods := len(list)
	if numPods == 0 {
		return nil, &platform.PodNotFoundError{Selector: selector}
	} else if numPods > 1 {
		return nil, fmt.Errorf("multiple Pods exist for the selector: %v. Only one must be present", selector)
	}

	podReport := list[0]
	var pod corev1.Pod
	pod.SetName(podReport.Name)
	pod.SetLabels(podReport.Labels)

	inspect, err := o.PodInspect(podReport.Name)
	if err != nil {
		return nil, err
	}
	if inspect.State == "Running" {
		pod.Status.Phase = corev1.PodRunning
	}
	return &pod, nil
}

func (o *PodmanCli) getPodsFromSelector(selector string) ([]ListPodsReport, error) {
	args := []string{"pod", "ps", "--format", "json"}
	selectorAsList := strings.Split(selector, ",")
	for _, s := range selectorAsList {
		args = append(args, "--filter=label="+s)
	}
	cmd := exec.Command(o.podmanCmd, args...)
	klog.V(3).Infof("executing %v", cmd.Args)
	out, err := cmd.Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		return nil, err
	}
	var list []ListPodsReport
	if err = json.Unmarshal(out, &list); err != nil {
		return nil, err
	}
	return list, nil
}
