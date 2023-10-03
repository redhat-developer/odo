package podman

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/platform"
)

// GetPodsMatchingSelector returns all pods matching the given label selector.
func (o *PodmanCli) GetPodsMatchingSelector(selector string) (*corev1.PodList, error) {
	podsReport, err := o.getPodsFromSelector(selector)
	if err != nil {
		return nil, err
	}

	var result corev1.PodList
	for _, podReport := range podsReport {
		pod, err := o.KubeGenerate(podReport.Name)
		if err != nil {
			// The pod has disappeared in the meantime, forget it
			continue
		}
		// We remove the podname- prefix from the container names as Podman adds this prefix
		// (to avoid colliding container names?)
		for c := range pod.Spec.Containers {
			container := &pod.Spec.Containers[c]
			prefix := pod.GetName() + "-"
			container.Name = strings.TrimPrefix(container.Name, prefix)
		}
		inspect, err := o.PodInspect(podReport.Name)
		if err != nil {
			// The pod has disappeared in the meantime, forget it
			continue
		}
		pod.Status.Phase = corev1.PodPhase(inspect.State)

		result.Items = append(result.Items, *pod)
	}
	return &result, nil
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
	// In podman, we return the pods, as there is no resource containing PodSpec
	return o.GetPodsMatchingSelector(selector)
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
	if inspect.State != "Running" {
		return nil, fmt.Errorf("a pod exists but is not in Running state. Current status=%v", inspect.State)
	}

	for _, container := range podReport.Containers {
		// Names of users containers are prefixed with pod name by podman
		if !strings.HasPrefix(container.Names, podReport.Name+"-") {
			continue
		}
		pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{
			Name: strings.TrimPrefix(container.Names, podReport.Name+"-"),
		})
	}
	return &pod, nil
}

func (o *PodmanCli) getPodsFromSelector(selector string) ([]ListPodsReport, error) {
	args := []string{"pod", "ps", "--format", "json"}
	selectorAsList := strings.Split(selector, ",")
	for _, s := range selectorAsList {
		args = append(args, "--filter=label="+s)
	}
	cmd := exec.Command(o.podmanCmd, append(o.containerRunGlobalExtraArgs, args...)...)
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

type podWatcher struct {
	stop   chan struct{}
	pods   map[string]struct{}
	events chan watch.Event
}

func (o *PodmanCli) PodWatcher(ctx context.Context, selector string) (watch.Interface, error) {

	watcher := podWatcher{
		stop:   make(chan struct{}),
		pods:   make(map[string]struct{}),
		events: make(chan watch.Event),
	}
	go watcher.watch(o.podmanCmd, o.containerRunGlobalExtraArgs)
	return watcher, nil
}

func (o podWatcher) watch(podmanCmd string, containerRunGlobalExtraArgs []string) {
	args := []string{"ps", "--quiet"}
	args = append(containerRunGlobalExtraArgs, args...)
	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-o.stop:
			return
		case <-ticker.C:
			cmd := exec.Command(podmanCmd, args...)
			out, err := cmd.Output()
			if err != nil {
				klog.V(4).Infof("error getting containers from podman: %s", err)
				continue
			}
			scanner := bufio.NewScanner(bytes.NewReader(out))
			currentPods := make(map[string]struct{})
			for scanner.Scan() {
				podName := scanner.Text()
				currentPods[podName] = struct{}{}
				if _, ok := o.pods[podName]; !ok {
					o.events <- watch.Event{
						Type: watch.Added,
					}
					o.pods[podName] = struct{}{}
				}
			}
			for p := range o.pods {
				if _, ok := currentPods[p]; !ok {
					o.events <- watch.Event{
						Type: watch.Deleted,
					}
					delete(o.pods, p)
				}
			}
		}
	}
}

func (o podWatcher) Stop() {
	o.stop <- struct{}{}
}

func (o podWatcher) ResultChan() <-chan watch.Event {
	return o.events
}
