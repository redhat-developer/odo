package logs

import (
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"

	odolabels "github.com/redhat-developer/odo/pkg/labels"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/platform"
)

type LogsClient struct {
	platformClient platform.Client
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

func NewLogsClient(platformClient platform.Client) *LogsClient {
	return &LogsClient{
		platformClient: platformClient,
	}
}

var _ Client = (*LogsClient)(nil)

func (o *LogsClient) GetLogsForMode(
	ctx context.Context,
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

	go o.getLogsForMode(ctx, events, mode, componentName, namespace, follow)
	return events, nil
}

func (o *LogsClient) getLogsForMode(
	ctx context.Context,
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
					containerLogs, err := o.platformClient.GetPodLogs(pod.Name, container.Name, follow)
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

	appname := odocontext.GetApplication(ctx)

	if mode == odolabels.ComponentDevMode || mode == odolabels.ComponentAnyMode {
		selector = odolabels.GetSelector(componentName, appname, odolabels.ComponentDevMode, false)
		err := o.getPodsForSelector(selector, namespace, podChan)
		if err != nil {
			errChan <- err
		}
	}
	if mode == odolabels.ComponentDeployMode || mode == odolabels.ComponentAnyMode {
		selector = odolabels.GetSelector(componentName, appname, odolabels.ComponentDeployMode, false)
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
	// set of unique Pods with Pod name as key; these are the Pods whose logs we want to get from the cluster
	pods := map[string]struct{}{}

	podList, err := o.platformClient.GetPodsMatchingSelector(selector)
	if err != nil {
		return err
	}
	for _, pod := range podList.Items {
		pods[pod.GetName()] = struct{}{}
	}

	// get all pods in the namespace
	podsInNs, err := o.platformClient.GetAllPodsInNamespaceMatchingSelector(selector, namespace)
	if err != nil {
		return err
	}

	for _, pod := range podsInNs.Items {
		if _, ok := pods[pod.GetName()]; ok {
			// Pod's logs have already been displayed to user
			continue
		}
		podList.Items = append(podList.Items, pod)
	}

	for _, pod := range podList.Items {
		podChan <- pod
	}

	return nil
}
