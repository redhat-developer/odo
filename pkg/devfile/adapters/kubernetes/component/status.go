package component

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"time"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/pkg/errors"
	"k8s.io/klog"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/machineoutput"
)

// KubernetesDeploymentStatus is a simplified representation of the component's cluster resources
type KubernetesDeploymentStatus struct {
	DeploymentUID types.UID
	ReplicaSetUID types.UID
	Pods          []*corev1.Pod
}

// KubernetesPodStatus is a representation of corev1.Pod, but only containing the fields we are interested in (for later marshalling to JSON)
type KubernetesPodStatus struct {
	Name           string
	UID            string
	Phase          string
	Labels         map[string]string
	StartTime      *time.Time
	Containers     []corev1.ContainerStatus
	InitContainers []corev1.ContainerStatus
}

// Find the pod for the component and convert to KubernetesDeploymentStatus
func (a Adapter) getDeploymentStatus() (*KubernetesDeploymentStatus, error) {

	// 1) Retrieve the deployment
	deployment, err := a.Client.GetKubeClient().GetOneDeployment(a.ComponentName, a.AppName)
	if err != nil {
		klog.V(4).Infof("Unable to retrieve deployment %s in %s ", a.ComponentName, a.Client.Namespace)
		return nil, err
	}

	if deployment == nil {
		return nil, errors.New("deployment status from Kubernetes API was nil")
	}

	deploymentUID := deployment.UID

	// 2) Retrieve the replica set that is owned by the deployment; if multiple, go with one with largest generation
	replicaSetList, err := a.Client.GetKubeClient().KubeClient.AppsV1().ReplicaSets(a.Client.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	matchingReplicaSets := []v1.ReplicaSet{}
	sort.Slice(replicaSetList.Items, func(i, j int) bool {
		iGen := replicaSetList.Items[i].Generation
		jGen := replicaSetList.Items[j].Generation

		// Sort descending by generation
		return iGen > jGen
	})

	// Locate the first matching replica, after above sort
outer:
	for _, replicaSet := range replicaSetList.Items {
		for _, ownerRef := range replicaSet.OwnerReferences {
			if ownerRef.UID == deploymentUID {
				matchingReplicaSets = append(matchingReplicaSets, replicaSet)
				break outer
			}
		}
	}

	if len(matchingReplicaSets) == 0 {
		return nil, errors.New("no replica sets found")
	}

	replicaSetUID := matchingReplicaSets[0].UID

	// 3) Retrieves the pods that are owned by the ReplicaSet and return
	podList, err := a.Client.GetKubeClient().KubeClient.CoreV1().Pods(a.Client.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	matchingPods := []*corev1.Pod{}
	for i, podItem := range podList.Items {
		for _, ownerRef := range podItem.OwnerReferences {

			if string(ownerRef.UID) == string(replicaSetUID) {
				matchingPods = append(matchingPods, &podList.Items[i])
			}
		}
	}
	result := KubernetesDeploymentStatus{}
	result.Pods = append(result.Pods, matchingPods...)

	result.DeploymentUID = deploymentUID
	result.ReplicaSetUID = replicaSetUID

	return &result, nil

}

// CreateKubernetesPodStatusFromPod extracts only the fields we are interested in from corev1.Pod
func CreateKubernetesPodStatusFromPod(pod corev1.Pod) KubernetesPodStatus {
	podStatus := KubernetesPodStatus{
		Name:           pod.Name,
		UID:            string(pod.UID),
		Phase:          string(pod.Status.Phase),
		Labels:         pod.Labels,
		InitContainers: []corev1.ContainerStatus{},
		Containers:     []corev1.ContainerStatus{},
	}

	if pod.Status.StartTime != nil {
		podStatus.StartTime = &pod.Status.StartTime.Time
	}

	podStatus.InitContainers = pod.Status.InitContainerStatuses

	podStatus.Containers = pod.Status.ContainerStatuses

	return podStatus

}

const (
	// SupervisordCheckInterval is the time we wait before we check the supervisord statuses each time, after the first call
	SupervisordCheckInterval = time.Duration(10) * time.Second
)

// StartSupervisordCtlStatusWatch kicks off a goroutine which calls 'supervisord ctl status' within every odo-managed container, every X seconds,
// and reports the result to the console.
func (a Adapter) StartSupervisordCtlStatusWatch() {

	watcher := newSupervisordStatusWatch(a.Logger())

	ticker := time.NewTicker(SupervisordCheckInterval)

	go func() {

		for {
			// On initial goroutine start, perform a query
			watcher.querySupervisordStatusFromContainers(a)
			<-ticker.C
		}

	}()

}

type supervisordStatusWatcher struct {
	// See 'createSupervisordStatusReconciler' for a description of the reconciler
	statusReconcilerChannel chan supervisordStatusEvent
}

func newSupervisordStatusWatch(loggingClient machineoutput.MachineEventLoggingClient) *supervisordStatusWatcher {
	inputChan := createSupervisordStatusReconciler(loggingClient)

	return &supervisordStatusWatcher{
		statusReconcilerChannel: inputChan,
	}
}

// createSupervisordStatusReconciler contains the status reconciler implementation.
// The reconciler receives (is sent) channel messages that contains the 'supervisord ctl status' values for each odo-managed container,
// with the result reported to the console.
func createSupervisordStatusReconciler(loggingClient machineoutput.MachineEventLoggingClient) chan supervisordStatusEvent {

	senderChannel := make(chan supervisordStatusEvent)

	go func() {
		// Map key: 'podUID:containerName' (within pod) -> list of statuses from 'supervisord ctl status'
		lastContainerStatus := map[string][]supervisordStatus{}

		for {

			event := <-senderChannel

			key := event.podUID + ":" + event.containerName

			previousStatus, hasLastContainerStatus := lastContainerStatus[key]
			lastContainerStatus[key] = event.status

			reportChange := false

			if hasLastContainerStatus {
				// If we saw a status for this container previously...
				if !supervisordStatusesEqual(previousStatus, event.status) {
					reportChange = true
				} else {
					reportChange = false
				}

			} else {
				// No status from the container previously...
				reportChange = true
			}

			entries := []machineoutput.SupervisordStatusEntry{}

			for _, status := range event.status {
				entries = append(entries, machineoutput.SupervisordStatusEntry{
					Program: status.program,
					Status:  status.status,
				})
			}

			loggingClient.SupervisordStatus(entries, machineoutput.TimestampNow())

			if reportChange {
				klog.V(4).Infof("Ccontainer %v status has changed - is: %v", event.containerName, event.status)
			}

		}

	}()

	return senderChannel
}

// querySupervisordStatusFromContainers locates the correct component's pod, and for each container within the pod queries the supervisord ctl status.
// The status results are sent to the reconciler.
func (sw *supervisordStatusWatcher) querySupervisordStatusFromContainers(a Adapter) {

	status, err := a.getDeploymentStatus()
	if err != nil {
		a.Logger().ReportError(errors.Wrap(err, "unable to retrieve container status"), machineoutput.TimestampNow())
		return
	}

	if status == nil {
		return
	}

	// Given a list of odo-managed pods, we want to find the newest; if there are multiple with the same age, then find the most
	// alive by container status.
	var podPhaseSortOrder = map[corev1.PodPhase]int{
		corev1.PodFailed:    0,
		corev1.PodSucceeded: 1,
		corev1.PodUnknown:   2,
		corev1.PodPending:   3,
		corev1.PodRunning:   4,
	}
	sort.Slice(status.Pods, func(i, j int) bool {

		iPod := status.Pods[i]
		jPod := status.Pods[j]

		iTime := iPod.CreationTimestamp.Time
		jTime := jPod.CreationTimestamp.Time

		if !jTime.Equal(iTime) {
			// Sort descending by creation timestamp
			return jTime.After(iTime)
		}

		// Next, sort descending to find the pod with most successful pod phase:
		// PodRunning > PodPending > PodUnknown > PodSucceeded > PodFailed
		return podPhaseSortOrder[jPod.Status.Phase] > podPhaseSortOrder[iPod.Status.Phase]
	})

	if len(status.Pods) < 1 {
		return
	}

	// Retrieve the first pod, which post-sort should be the most recent and most alive
	pod := status.Pods[0]

	debugCommand, err := common.GetDebugCommand(a.Devfile.Data, a.devfileDebugCmd)
	if err != nil {
		a.Logger().ReportError(errors.Wrap(err, "unable to retrieve debug command"), machineoutput.TimestampNow())
		return
	}

	runCommand, err := common.GetRunCommand(a.Devfile.Data, a.devfileRunCmd)
	if err != nil {
		a.Logger().ReportError(errors.Wrap(err, "unable to retrieve run command"), machineoutput.TimestampNow())
		return
	}

	// For each of the containers, retrieve the status of the tasks and send that status back to the status reconciler
	for _, container := range pod.Status.ContainerStatuses {

		if (runCommand.Exec != nil && container.Name == runCommand.Exec.Component) || (debugCommand.Exec != nil && container.Name == debugCommand.Exec.Component) {
			status := getSupervisordStatusInContainer(pod.Name, container.Name, a)

			sw.statusReconcilerChannel <- supervisordStatusEvent{
				containerName: container.Name,
				status:        status,
				podUID:        string(pod.UID),
			}
		}
	}
}

// supervisordStatusesEqual is a simple comparison of []supervisord that ignores slice element order
func supervisordStatusesEqual(one []supervisordStatus, two []supervisordStatus) bool {
	if len(one) != len(two) {
		return false
	}

	for _, oneVal := range one {

		match := false
		for _, twoVal := range two {

			if reflect.DeepEqual(oneVal, twoVal) {
				match = true
			}
		}
		if !match {
			return false
		}
	}
	return true
}

// getSupervisordStatusInContainer executes 'supervisord ctl status' within the pod and container, parses the output,
// and returns the status for the container
func getSupervisordStatusInContainer(podName string, containerName string, a Adapter) []supervisordStatus {

	command := []string{common.SupervisordBinaryPath, common.SupervisordCtlSubCommand, "status"}
	compInfo := common.ComponentInfo{
		ContainerName: containerName,
		PodName:       podName,
	}

	stdoutWriter, stdoutOutputChannel := common.CreateConsoleOutputWriterAndChannel()
	stderrWriter, stderrOutputChannel := common.CreateConsoleOutputWriterAndChannel()

	err := common.ExecuteCommand(&a, compInfo, command, false, stdoutWriter, stderrWriter)

	// Close the writer and wait the console output
	stdoutWriter.Close()
	consoleResult := <-stdoutOutputChannel

	stderrWriter.Close()
	consoleStderrResult := <-stderrOutputChannel

	if err != nil {
		a.Logger().ReportError(errors.Wrapf(err, "unable to execute command on %s within container %s, %v, output: %v %v", podName, containerName, err, consoleResult, consoleStderrResult), machineoutput.TimestampNow())
		return nil
	}

	result := []supervisordStatus{}

	for _, line := range consoleResult {

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		result = append(result, supervisordStatus{program: fields[0], status: fields[1]})
	}

	return result
}

// supervisordStatus corresponds to the statuses reported by 'supervisord ctl status', example:
// - debugrun                         STOPPED
// - devrun                           RUNNING   pid 5640, uptime 11 days, 21:56:20
// Only the first and second fields are included (no pod, uptime, etc)
type supervisordStatus struct {
	program string
	status  string
}

// All statuses seen within the container
type supervisordStatusEvent struct {
	containerName string
	podUID        string
	status        []supervisordStatus
}
