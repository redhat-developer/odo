package logs

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/fatih/color"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/log"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/platform"
)

type LogsClient struct {
	platformClient platform.Client
}

type ContainerLogs struct {
	PodName       string
	ContainerName string
	Logs          io.ReadCloser
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

func (o *LogsClient) DisplayLogs(
	ctx context.Context,
	mode string,
	componentName string,
	namespace string,
	follow bool,
	out io.Writer,
) error {
	events, err := o.GetLogsForMode(
		ctx,
		mode,
		componentName,
		namespace,
		follow,
	)
	if err != nil {
		return err
	}

	uniqueContainerNames := map[string]struct{}{}
	var goroutines struct{ count int64 } // keep a track of running goroutines so that we don't exit prematurely
	errChan := make(chan error)          // errors are put on this channel
	var mu sync.Mutex

	displayedLogs := map[string]struct{}{}
	for {
		select {
		case containerLogs := <-events.Logs:
			podContainerName := fmt.Sprintf("%s-%s", containerLogs.PodName, containerLogs.ContainerName)
			if _, ok := displayedLogs[podContainerName]; ok {
				continue
			}
			displayedLogs[podContainerName] = struct{}{}

			uniqueName := getUniqueContainerName(containerLogs.ContainerName, uniqueContainerNames)
			uniqueContainerNames[uniqueName] = struct{}{}
			colour := log.ColorPicker()
			logs := containerLogs.Logs

			func() {
				mu.Lock()
				defer mu.Unlock()
				color.Set(colour)
				defer color.Unset()
				help := ""
				if uniqueName != containerLogs.ContainerName {
					help = fmt.Sprintf(" (%s)", uniqueName)
				}
				_, err = fmt.Fprintf(out, "--> Logs for %s / %s%s\n", containerLogs.PodName, containerLogs.ContainerName, help)
				if err != nil {
					errChan <- err
				}
			}()

			if follow {
				atomic.AddInt64(&goroutines.count, 1)
				go func(out io.Writer) {
					defer func() {
						atomic.AddInt64(&goroutines.count, -1)
					}()
					err = printLogs(uniqueName, logs, out, colour, &mu)
					if err != nil {
						errChan <- err
					}
					delete(displayedLogs, podContainerName)
					events.Done <- struct{}{}
				}(out)
			} else {
				err = printLogs(uniqueName, logs, out, colour, &mu)
				if err != nil {
					return err
				}
			}
		case err = <-errChan:
			return err
		case err = <-events.Err:
			return err
		case <-events.Done:
			if !follow && goroutines.count == 0 {
				if len(uniqueContainerNames) == 0 {
					// This will be the case when:
					// 1. user specifies --dev flag, but the component's running in Deploy mode
					// 2. user specified --deploy flag, but the component's running in Dev mode
					// 3. user passes no flag, but component is running in neither Dev nor Deploy mode
					fmt.Fprintf(out, "no containers running in the specified mode for the component %q\n", componentName)
				}
				return nil
			}
		}
	}
}

func getUniqueContainerName(name string, uniqueNames map[string]struct{}) string {
	if _, ok := uniqueNames[name]; ok {
		// name already present in uniqueNames; find another name
		// first check if last character in name is a number; if so increment it, else append name with [1]
		var numStr string
		var last int
		var err error

		split := strings.Split(name, "[")
		if len(split) == 2 {
			numStr = strings.Trim(split[1], "]")
			last, err = strconv.Atoi(numStr)
			if err != nil {
				return ""
			}
			last++
		} else {
			last = 1
		}
		name = fmt.Sprintf("%s[%d]", split[0], last)
		return getUniqueContainerName(name, uniqueNames)
	}
	return name
}

// printLogs prints the logs of the containers with container name prefixed to the log message
func printLogs(containerName string, rd io.ReadCloser, out io.Writer, colour color.Attribute, mu *sync.Mutex) error {
	scanner := bufio.NewScanner(rd)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		err := func() error {
			mu.Lock()
			defer mu.Unlock()
			color.Set(colour)
			defer color.Unset()

			_, err := fmt.Fprintln(out, containerName+": "+line)
			return err
		}()
		if err != nil {
			return err
		}
	}

	return nil
}

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
					events.Logs <- ContainerLogs{
						PodName:       pod.GetName(),
						ContainerName: container.Name,
						Logs:          containerLogs,
					}
				}
			case err := <-errChan:
				events.Err <- err
			case <-doneChan:
				events.Done <- struct{}{}
			}
		}
	}()

	appname := odocontext.GetApplication(ctx)

	getPods := func() error {
		if mode == odolabels.ComponentDevMode || mode == odolabels.ComponentAnyMode {
			selector = odolabels.GetSelector(componentName, appname, odolabels.ComponentDevMode, false)
			err := o.getPodsForSelector(selector, namespace, podChan)
			if err != nil {
				return err
			}
		}
		if mode == odolabels.ComponentDeployMode || mode == odolabels.ComponentAnyMode {
			selector = odolabels.GetSelector(componentName, appname, odolabels.ComponentDeployMode, false)
			err := o.getPodsForSelector(selector, namespace, podChan)
			if err != nil {
				return err
			}
		}
		return nil
	}

	err := getPods()
	if err != nil {
		errChan <- err
	}

	if follow {
		podWatcher, err := o.platformClient.PodWatcher(ctx, "")
		if err != nil {
			errChan <- err
		}
		for ev := range podWatcher.ResultChan() {
			switch ev.Type {
			case watch.Added, watch.Modified:
				err = getPods()
				if err != nil {
					errChan <- err
				}
			}
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
		if pod.Status.Phase == "Running" {
			pods[pod.GetName()] = struct{}{}
		}
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
		if pod.Status.Phase == "Running" {
			podList.Items = append(podList.Items, pod)
		}
	}

	for _, pod := range podList.Items {
		if pod.Status.Phase == "Running" {
			podChan <- pod
		}
	}

	return nil
}
