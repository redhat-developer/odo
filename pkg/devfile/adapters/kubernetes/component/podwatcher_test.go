package component

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/devfile/library/pkg/devfile/parser/data"

	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	adaptersCommon "github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/occlient"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStatusReconciler(t *testing.T) {
	componentName := "my-component"

	tests := []struct {
		name              string
		pre               []testReconcilerEntry
		expectedPreEvents int
		post              []testReconcilerEntry
		successFn         func(lfo *logFuncOutput) string
	}{
		{
			name:              "a new pod should trigger a status update",
			pre:               []testReconcilerEntry{},
			expectedPreEvents: 0,
			post: []testReconcilerEntry{
				{
					pods: []*corev1.Pod{
						createFakePod(componentName, componentName, nil),
					},
				},
			},
			successFn: func(lfo *logFuncOutput) string {

				latestPodStatus := lfo.getMostRecentKubernetesPodStatus()
				if latestPodStatus == nil {
					return "pod not found"
				}
				if len(latestPodStatus.Pods) != 1 {
					return fmt.Sprintf("unexpected pod size, %v", lfo.debugSprintAll())
				}

				if latestPodStatus.Pods[0].Name != "my-component" {
					return fmt.Sprintf("mismatching component %v", lfo.debugSprintAll())
				}

				return ""
			},
		},
		{
			name: "if a pod is deleted, trigger a status update",
			pre: []testReconcilerEntry{
				{
					pods: []*corev1.Pod{createFakePod(componentName, componentName, func(pod *corev1.Pod) {
						pod.UID = "one"
					}),
					},
				},
			},

			expectedPreEvents: 1,
			post: []testReconcilerEntry{
				{
					pods: []*corev1.Pod{createFakePod(componentName, componentName, func(pod *corev1.Pod) {
						pod.UID = "one"
					}),
					},
					isDeleteEventFromWatch: true,
				},
			},
			successFn: func(lfo *logFuncOutput) string {
				latestPodStatus := lfo.getMostRecentKubernetesPodStatus()
				if latestPodStatus == nil {
					return "pod not found"
				}

				if len(latestPodStatus.Pods) != 0 {
					return fmt.Sprintf("Unexpected number of pods: %v", lfo.debugSprintAll())
				}

				return ""
			},
		},
		{
			name: "if a pod is updated, trigger a status update",
			pre: []testReconcilerEntry{
				{
					pods: []*corev1.Pod{
						createFakePod(componentName, componentName, func(pod *corev1.Pod) {
							pod.UID = "one"
							pod.Status.Phase = corev1.PodPending
						}),
					},
				},
			},

			expectedPreEvents: 1,
			post: []testReconcilerEntry{
				{
					pods: []*corev1.Pod{
						createFakePod(componentName, componentName, func(pod *corev1.Pod) {
							pod.UID = "one"
							pod.Status.Phase = corev1.PodRunning
						}),
					},
				},
			},

			successFn: func(lfo *logFuncOutput) string {
				latestPodStatus := lfo.getMostRecentKubernetesPodStatus()
				if latestPodStatus == nil {
					return "pod not found"
				}

				if len(latestPodStatus.Pods) != 1 {
					return fmt.Sprintf("unexpected pod size, %v", lfo.debugSprintAll())
				}

				if latestPodStatus.Pods[0].Name != "my-component" {
					return fmt.Sprintf("mismatching component, %v", lfo.debugSprintAll())
				}

				if latestPodStatus.Pods[0].Phase != string(corev1.PodRunning) {
					return fmt.Sprintf("unexpected pod phase, %v", lfo.debugSprintAll())
				}

				return ""

			},
		},
		{
			name: "if a pod fails and is replaced by another, but both temporarily exist together",
			pre: []testReconcilerEntry{
				{
					pods: []*corev1.Pod{
						createFakePod(componentName, componentName, func(pod *corev1.Pod) {
							pod.UID = "one"
							pod.Status.Phase = corev1.PodPending
						}),
					},
				},
			},
			expectedPreEvents: 1,
			post: []testReconcilerEntry{
				{
					pods: []*corev1.Pod{
						createFakePod(componentName, componentName, func(pod *corev1.Pod) {
							pod.UID = "one"
							pod.Status.Phase = corev1.PodFailed
						}),
					},
				},
				{
					pods: []*corev1.Pod{
						createFakePod(componentName, componentName, func(pod *corev1.Pod) {
							pod.UID = "two"
							pod.Status.Phase = corev1.PodRunning
						}),
					},
				},
			},

			successFn: func(lfo *logFuncOutput) string {
				latestPodStatus := lfo.getMostRecentKubernetesPodStatus()
				if latestPodStatus == nil {
					return "pod not found"
				}

				if len(latestPodStatus.Pods) != 2 {
					return fmt.Sprintf("unexpected pod size, %v", lfo.debugSprintAll())
				}

				for _, pod := range latestPodStatus.Pods {

					if pod.Name != "my-component" {
						return fmt.Sprintf("mismatching component, %v", lfo.debugSprintAll())
					}

					if pod.UID == "one" {
						if pod.Phase != string(corev1.PodFailed) {
							return fmt.Sprintf("unexpected pod phase, %v", lfo.debugSprintAll())
						}
					}

					if pod.UID == "two" {
						if pod.Phase != string(corev1.PodRunning) {
							return fmt.Sprintf("unexpected pod phase, %v", lfo.debugSprintAll())
						}
					}

				}

				return ""

			},
		},
		{
			name: "if a pod fails, and is fully replaced (one and new pod don't co-exist at the same time)",

			pre: []testReconcilerEntry{
				{
					pods: []*corev1.Pod{
						createFakePod(componentName, componentName, func(pod *corev1.Pod) {
							pod.UID = "one"
							pod.Status.Phase = corev1.PodRunning
						}),
					},
				},
			},

			expectedPreEvents: 1,
			post: []testReconcilerEntry{
				{
					pods: []*corev1.Pod{
						createFakePod(componentName, componentName, func(pod *corev1.Pod) {
							pod.UID = "one"
							pod.Status.Phase = corev1.PodFailed
						}),
					},
					isDeleteEventFromWatch: true,
				},
				{
					pods: []*corev1.Pod{
						createFakePod(componentName, componentName, func(pod *corev1.Pod) {
							pod.UID = "two"
							pod.Status.Phase = corev1.PodRunning
						}),
					},
				},
			},
			successFn: func(lfo *logFuncOutput) string {
				latestPodStatus := lfo.getMostRecentKubernetesPodStatus()
				if latestPodStatus == nil {
					return "pod not found"
				}

				if len(latestPodStatus.Pods) != 1 {
					return fmt.Sprintf("unexpected pod size, %v", lfo.debugSprintAll())
				}

				if latestPodStatus.Pods[0].Name != "my-component" {
					return fmt.Sprintf("mismatching component, %v", lfo.debugSprintAll())
				}

				if latestPodStatus.Pods[0].UID != "two" {
					return fmt.Sprintf("unexpected pod UID, %v", lfo.debugSprintAll())
				}
				return ""

			},
		},

		{
			name:              "no changes should trigger no events",
			pre:               []testReconcilerEntry{},
			expectedPreEvents: 0,
			post:              []testReconcilerEntry{},
			successFn: func(lfo *logFuncOutput) string {
				time.Sleep(5 * time.Second)

				if lfo.listSize() > 0 {
					return fmt.Sprintf("unexpected events in output %v", lfo.debugSprintAll())
				}

				return ""
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: componentName,
				Devfile:       devObj,
			}

			fkclient, _ := occlient.FakeNew()

			adapter := New(adapterCtx, *fkclient)

			lfo := logFuncOutput{}
			adapter.GenericAdapter.SetLogger(machineoutput.NewConsoleMachineEventLoggingClientWithFunction(lfo.logFunc))

			reconcilerChannel := createStatusReconciler(&adapter)

			// Initialize with an empty list
			reconcilerChannel <- statusReconcilerChannelEntry{
				pods:                   []*corev1.Pod{},
				err:                    nil,
				isCompleteListOfPods:   true,
				isDeleteEventFromWatch: false,
			}

			for _, fauxReconcilerEntry := range tt.pre {
				// Send the initial simulated cluster status before the test runs
				reconcilerChannel <- statusReconcilerChannelEntry{
					pods:                   fauxReconcilerEntry.pods,
					err:                    nil,
					isCompleteListOfPods:   fauxReconcilerEntry.isCompleteListOfPods,
					isDeleteEventFromWatch: fauxReconcilerEntry.isDeleteEventFromWatch,
				}

			}

			// Wait for the expected number of events that will be generated by sending the initial faux cluster status
			expireTime := time.Now().Add(5 * time.Second)
			for lfo.listSize()-tt.expectedPreEvents != 0 {
				time.Sleep(20 * time.Millisecond)

				if time.Now().After(expireTime) {
					t.Fatalf("unexpected number of pre events: %v", lfo.debugSprintAll())
				}
			}

			// Clear the expected events
			lfo.clearList()

			for _, fauxReconcilerEntry := range tt.post {
				// Send the test's simulated cluster status
				reconcilerChannel <- statusReconcilerChannelEntry{
					pods:                   fauxReconcilerEntry.pods,
					err:                    nil,
					isCompleteListOfPods:   fauxReconcilerEntry.isCompleteListOfPods,
					isDeleteEventFromWatch: fauxReconcilerEntry.isDeleteEventFromWatch,
				}
			}

			// Wait up to 10 seconds for the test to signal success (an empty string, indicating no errors)
			expireTime = time.Now().Add(10 * time.Second)
			mostRecentError := ""
			for {
				failureReason := tt.successFn(&lfo)

				mostRecentError = failureReason

				if failureReason == "" {
					break
				}

				if time.Now().After(expireTime) {
					break
				}
			}

			if mostRecentError != "" {
				t.Fatal(mostRecentError)
			}

			if lfo.errorOccurred != nil {
				t.Fatalf("error occurred during test case run %v", lfo.errorOccurred)
			}

		})
	}

}

// Simulate a channel message sent to the status reconciler. See 'statusReconcilerChannelEntry' for field details
type testReconcilerEntry struct {
	pods []*corev1.Pod

	isCompleteListOfPods bool

	isDeleteEventFromWatch bool
}

// getMostRecentKubernetesPodStatus is a test convenience method to retrieve the most recent pod status
func (lfo *logFuncOutput) getMostRecentKubernetesPodStatus() *machineoutput.KubernetesPodStatus {

	lfo.listMutex.Lock()
	defer lfo.listMutex.Unlock()

	var podStatus *machineoutput.KubernetesPodStatus

	for _, entry := range lfo.jsonList {

		if entry.GetType() == machineoutput.TypeKubernetesPodStatus {
			podStatus = entry.(*machineoutput.KubernetesPodStatus)
		}
	}

	return podStatus
}

// listSize is simple thread-safe wrapper around list
func (lfo *logFuncOutput) listSize() int {
	lfo.listMutex.Lock()
	defer lfo.listMutex.Unlock()

	return len(lfo.jsonList)
}

// debugSprintAll returns a list of all machine readable JSON events that have been output thus far
func (lfo *logFuncOutput) debugSprintAll() string {

	lfo.listMutex.Lock()
	defer lfo.listMutex.Unlock()

	result := ""

	for _, entry := range lfo.jsonList {
		jsonVal, err := json.Marshal(entry)
		if err != nil {
			lfo.errorOccurred = err
			return fmt.Sprint(err)
		}
		result += string(jsonVal)
	}

	return result
}

// clearList clears the internal list of received machine readable JSON events
func (lfo *logFuncOutput) clearList() {
	lfo.listMutex.Lock()
	defer lfo.listMutex.Unlock()

	lfo.jsonList = []machineoutput.MachineEventLogEntry{}
}

// Any machine readable JSON events that are output by odo are passed to this function, and this function
// adds them to an internal list, for test verification
func (lfo *logFuncOutput) logFunc(wrapper machineoutput.MachineEventWrapper) {

	lfo.listMutex.Lock()
	defer lfo.listMutex.Unlock()

	json, err := wrapper.GetEntry()
	if err != nil {
		lfo.errorOccurred = err
		return
	}

	machineoutput.OutputSuccessUnindented(wrapper)

	lfo.jsonList = append(lfo.jsonList, json)
}

type logFuncOutput struct {
	jsonList      []machineoutput.MachineEventLogEntry
	listMutex     sync.Mutex
	errorOccurred error
}

func createFakePod(componentName, podName string, fn func(*corev1.Pod)) *corev1.Pod {
	fakePod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
			Labels: map[string]string{
				"component": componentName,
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	if fn != nil {
		fn(fakePod)
	}

	return fakePod
}
