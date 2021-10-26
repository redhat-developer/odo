package component

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/openshift/odo/v2/pkg/machineoutput"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog"
)

const (
	// KubernetesResourceFailureInterval is the time between attempts to acquire needed k8s resources
	KubernetesResourceFailureInterval = time.Duration(5) * time.Second
)

// podWatcher is responsible for watching for changes to odo-managed Pods, and reporting those changes to the console, as used by the status command
type podWatcher struct {
	adapter              *Adapter
	statusReconcilerChan chan statusReconcilerChannelEntry
}

// StartContainerStatusWatch outputs Kubernetes pod/container status changes to the console, as used by the status command
func (a Adapter) StartContainerStatusWatch() {

	pw := newPodWatcher(&a)
	pw.startPodWatcher()

}

func newPodWatcher(adapter *Adapter) *podWatcher {
	return &podWatcher{
		adapter:              adapter,
		statusReconcilerChan: createStatusReconciler(adapter),
	}
}

func (pw *podWatcher) startPodWatcher() {
	pw.startWatchThread(pw.adapter)
}

// statusReconcilerChannelEntry is the gochannel message sent from the Watcher to the status reconciler
type statusReconcilerChannelEntry struct {

	// If isCompleteListOfPods is true: a list of all component pods in the workspace
	// If isCompleteListOfPods is false: a single component pod in the workspace
	pods []*corev1.Pod

	err error

	// isCompleteListOfPods is true if the pods came from getLatestContainerStatus(), false otherwise
	isCompleteListOfPods bool

	// isDeleteEventFromWatch is true if watch.Deleted event from the watch, false otherwise
	isDeleteEventFromWatch bool

	// watchThreadRestarted is true if the watch thread died (for example, due to losing network connection) and had to be reestablished
	isWatchThreadRestarted bool
}

// getLatestContainerStatus returns a KubernetesDeploymentStatus for the given component; this function blocks until it is available
func getLatestContainerStatus(adapter *Adapter) *KubernetesDeploymentStatus {

	// Keep trying to acquire the ReplicaSet and DeploymentSet of the component, so that we can reliably find its pods
	for {
		containerStatus, err := adapter.getDeploymentStatus()
		if err == nil {

			if containerStatus.DeploymentUID == "" || containerStatus.ReplicaSetUID == "" {
				adapter.Logger().ReportError(fmt.Errorf("unable to retrieve component deployment and replica set, trying again in a few moments"), machineoutput.TimestampNow())
				time.Sleep(KubernetesResourceFailureInterval)
				continue
			}

			return containerStatus
		}

		adapter.Logger().ReportError(errors.Wrapf(err, "unable to retrieve component deployment and replica set, trying again in a few moments"), machineoutput.TimestampNow())
		time.Sleep(KubernetesResourceFailureInterval)
	}

}

func (pw *podWatcher) startWatchThread(adapter *Adapter) {

	// Kick off the goroutine then return execution
	go func() {

		watchAttempts := 1

		var w watch.Interface = nil
		for {

			klog.V(4).Infof("Attempting to acquire watch, attempt #%d", watchAttempts)

			var err error
			w, err = adapter.Client.GetKubeClient().KubeClient.CoreV1().Pods(adapter.Client.Namespace).Watch(context.TODO(), metav1.ListOptions{})

			if err != nil || w == nil {

				if err != nil {
					adapter.Logger().ReportError(err, machineoutput.TimestampNow())
				}

				klog.V(4).Infof("Unable to establish watch, trying again in a few moments seconds. Error was:  %v", err)

				time.Sleep(KubernetesResourceFailureInterval)
				watchAttempts++
			} else {
				// Success!
				break
			}
		}

		klog.V(4).Infof("Watch is successfully established.")

		kubeContainerStatus := getLatestContainerStatus(adapter)

		// After the watch is established, provide the reconciler with a list of all the current pods in the namespace (not just delta), so that
		// old pods may be deleted from the reconciler (eg those pods that were deleted in the namespace while the watch was dead).
		// (This prevents a race condition where pods deleted during a watch outage might be missed forever).
		pw.statusReconcilerChan <- statusReconcilerChannelEntry{
			pods:                   kubeContainerStatus.Pods,
			isCompleteListOfPods:   true,
			isDeleteEventFromWatch: false,
			err:                    nil,
		}

		// We have succesfully established the watch, so kick off the watch event listener
		go pw.watchEventListener(w, kubeContainerStatus.ReplicaSetUID)

	}()

}

// This function runs in a goroutine for each watch. This goroutine exits if the watch dies (for example due to network disconnect),
// at which point the watch acquisition process begins again.
func (pw *podWatcher) watchEventListener(w watch.Interface, replicaSetUID types.UID) {
	for {

		// Retrieve watch event
		entry := <-w.ResultChan()

		// Restart the watch acquisition process on death, then exit
		if entry.Object == nil && entry.Type == "" {
			klog.V(4).Infof("Watch has died; initiating re-establish.")
			pw.statusReconcilerChan <- statusReconcilerChannelEntry{
				isWatchThreadRestarted: true,
			}
			pw.startWatchThread(pw.adapter)
			return
		}

		// We only care about watch events that are related to Pods
		if pod, ok := entry.Object.(*corev1.Pod); ok && pod != nil {

			// Look for pods that are owned by the replicaset of our deployment
			ownerRefMatches := false
			for _, ownerRef := range pod.OwnerReferences {
				if ownerRef.UID == replicaSetUID {
					ownerRefMatches = true
					break
				}
			}
			if !ownerRefMatches {
				continue
			}

			// We located the component pod, so now pass it to our status reconciler to report to the console (if required)
			pw.statusReconcilerChan <- statusReconcilerChannelEntry{
				pods:                   []*corev1.Pod{pod},
				err:                    nil,
				isCompleteListOfPods:   false, // only a delta
				isDeleteEventFromWatch: entry.Type == watch.Deleted,
			}

		}
	}
}

// createStatusReconciler kicks off a goroutine which receives messages containing updates to odo-managed k8s Pod resources.
// For each message received, this function must determine if that resources has changed (in a way that we care about), and
// if so, report that as a change event.
func createStatusReconciler(adapter *Adapter) chan statusReconcilerChannelEntry {

	senderChannel := make(chan statusReconcilerChannelEntry)

	go func() {

		// This map is the single source of truth re: what odo expects the cluster namespace to look like; when
		// new events are received that contain pod data that differs from this, the user should be informed of the delta
		// (and this 'truth' should be updated.)
		//
		// Map key is pod UID
		mostRecentPodStatus := map[string]*KubernetesPodStatus{}

		for {

			entry := <-senderChannel

			if entry.isWatchThreadRestarted {
				// On network disconnect, clear the status map
				mostRecentPodStatus = map[string]*KubernetesPodStatus{}
			}

			if entry.err != nil {
				adapter.Logger().ReportError(entry.err, machineoutput.TimestampNow())
				klog.V(4).Infof("Error received on status reconciler channel %v", entry.err)
				continue
			}

			if entry.pods == nil {
				continue
			}

			// Map key is pod UID (we don't use the map value)
			entryPodUIDs := map[string]string{}
			for _, pod := range entry.pods {
				entryPodUIDs[string(pod.UID)] = string(pod.UID)
			}

			changeDetected := false

			// This section of the algorithm only works if the entry was from a podlist (which contains the full list
			// of all pods that exist in the namespace), rather than the watch (which contains only one pod in
			// the namespace.)
			if entry.isCompleteListOfPods {
				// Detect if there exists a UID in mostRecentPodStatus that is not in entry; if so, one or more previous
				// pods have disappeared, so set changeDetected to true.
				for mostRecentPodUID := range mostRecentPodStatus {
					if _, exists := entryPodUIDs[mostRecentPodUID]; !exists {
						klog.V(4).Infof("Status change detected: Could not find previous pod %s in most recent pod status", mostRecentPodUID)
						delete(mostRecentPodStatus, mostRecentPodUID)
						changeDetected = true
					}
				}
			}

			if !changeDetected {

				// For each pod we received a status for, determine if it is a change, and if so, update mostRecentPodStatus
				for _, pod := range entry.pods {
					podVal := CreateKubernetesPodStatusFromPod(*pod)

					if entry.isDeleteEventFromWatch {
						delete(mostRecentPodStatus, string(pod.UID))
						klog.V(4).Infof("Removing deleted pod %s", pod.UID)
						changeDetected = true
						continue
					}

					// If a pod exists in the new pod status, that we have not seen before, then a change is detected.
					prevValue, exists := mostRecentPodStatus[string(pod.UID)]
					if !exists {
						mostRecentPodStatus[string(pod.UID)] = &podVal
						klog.V(4).Infof("Adding new pod to most recent pod status %s", pod.UID)
						changeDetected = true

					} else {
						// If the pod exists in both the old and new status, then do a deep comparison
						areEqual := areEqual(&podVal, prevValue)
						if areEqual != "" {
							mostRecentPodStatus[string(pod.UID)] = &podVal
							klog.V(4).Infof("Pod value %s has changed:  %s", pod.UID, areEqual)
							changeDetected = true
						}
					}
				}
			}

			// On change: output all pods (our full knowledge of the odo-managed components in the namespace) as a single JSON event
			if changeDetected {

				podStatuses := []machineoutput.KubernetesPodStatusEntry{}

				for _, v := range mostRecentPodStatus {

					startTime := ""
					if v.StartTime != nil {
						startTime = machineoutput.FormatTime(*v.StartTime)
					}

					podStatuses = append(podStatuses, machineoutput.KubernetesPodStatusEntry{
						Name:           v.Name,
						Containers:     v.Containers,
						InitContainers: v.InitContainers,
						Labels:         v.Labels,
						Phase:          v.Phase,
						UID:            v.UID,
						StartTime:      startTime,
					})
				}

				adapter.Logger().KubernetesPodStatus(podStatuses, machineoutput.TimestampNow())
			}
		}
	}()

	return senderChannel
}

// areEqual compares two KubernetesPodStatus and returns a non-empty string if the two are not equivalent.
// Note: returned strings are for logging/debug purposes only.
func areEqual(one *KubernetesPodStatus, two *KubernetesPodStatus) string {

	if one.UID != two.UID {
		return fmt.Sprintf("UIDs differ %s %s", one.UID, two.UID)
	}

	if one.Name != two.Name {
		return fmt.Sprintf("Names differ %s %s", one.Name, two.Name)
	}

	if !reflect.DeepEqual(one.StartTime, two.StartTime) {
		return fmt.Sprintf("Start times differ %v %v", one.StartTime, two.StartTime)
	}

	if one.Phase != two.Phase {
		return fmt.Sprintf("Pod phase differs %s %s", one.Phase, two.Phase)
	}

	if !reflect.DeepEqual(one.Labels, two.Labels) {
		return fmt.Sprintf("Labels differ %v %v", one.Labels, two.Labels)
	}

	initContainerComparison := compareCoreContainerStatusList(one.InitContainers, two.InitContainers)
	if initContainerComparison != "" {
		return fmt.Sprintf("Init containers differ: %s", initContainerComparison)
	}

	containerComparison := compareCoreContainerStatusList(one.Containers, two.Containers)
	if containerComparison != "" {
		return fmt.Sprintf("Containers differ %s", containerComparison)
	}

	return ""
}

// compareCoreContainerStatusList compares two ContainerStatus arrays and returns a non-empty string if the two are not equivalent.
// Note: returned strings are for logging/debug purposes only.
func compareCoreContainerStatusList(oneParam []corev1.ContainerStatus, twoParam []corev1.ContainerStatus) string {

	// One-way list compare, using container name to identify individual entries
	compareFunc := func(paramA []corev1.ContainerStatus, paramB []corev1.ContainerStatus) string {

		// key: container name
		oneMap := map[string]*corev1.ContainerStatus{}

		// Populate oneMap
		for index, one := range paramA {
			oneMap[one.Name] = &paramA[index]
		}

		// Iterate through paramB and compare with the corresponding container name in paramA
		for index, two := range paramB {

			oneEntry, exists := oneMap[two.Name]

			// If an entry is present in two but not one
			if !exists || oneEntry == nil {
				return fmt.Sprintf("Container with id %s was present in one state but not the other", two.Name)
			}

			comparison := areCoreContainerStatusesEqual(oneEntry, &paramB[index])

			if comparison != "" {
				return comparison
			}
		}

		return ""
	}

	// Since compareFunc is unidirectional, we do it twice
	result := compareFunc(oneParam, twoParam)
	if result != "" {
		return result
	}

	result = compareFunc(twoParam, oneParam)
	return result

}

// areCoreContainerStatusesEqual compares two ContainerStatus and returns a non-empty string if the two are not equivalent.
// Note: returned strings are for logging/debug purposes only.
func areCoreContainerStatusesEqual(one *corev1.ContainerStatus, two *corev1.ContainerStatus) string {

	if one.Name != two.Name {
		return fmt.Sprintf("Core status names differ [%s] [%s]", one.Name, two.Name)
	}

	if one.ContainerID != two.ContainerID {
		return fmt.Sprintf("Core status container IDs differ: [%s] [%s]", one.ContainerID, two.ContainerID)
	}

	compareStates := compareCoreContainerState(one.State, two.State)
	if compareStates != "" {
		return fmt.Sprintf("Core status states differ %s", compareStates)
	}

	return ""
}

// compareCoreContainerState compares two ContainerState and returns a non-empty string if the two are not equivalent.
// Note: returned strings are for logging/debug purposes only.
func compareCoreContainerState(oneParam corev1.ContainerState, twoParam corev1.ContainerState) string {

	// At present, we only compare the state, and not the state contents, so convert the state to a string and
	// discard the other information.
	toString := func(one corev1.ContainerState) string {
		if one.Running != nil {
			return "Running"
		}

		if one.Terminated != nil {
			return "Terminated"
		}

		if one.Waiting != nil {
			return "Waiting"
		}

		return ""
	}

	oneParamState := toString(oneParam)
	twoParamState := toString(twoParam)

	if oneParamState != twoParamState {
		return "Core container states different: " + oneParamState + " " + twoParamState
	}

	return ""

}
