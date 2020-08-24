package component

import (
	"reflect"
	"strings"
	"time"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/docker/utils"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

const (
	// ContainerCheckInterval is the time we wait before we check the container statuses each time, after the first call
	ContainerCheckInterval = time.Duration(10) * time.Second

	// SupervisordCheckInterval is the time we call supervisord ctl status, after the first call
	SupervisordCheckInterval = time.Duration(10) * time.Second
)

// StartContainerStatusWatch outputs the state of every component container every X seconds; after the first, only
// state transitions will be outputted (for example, running -> stopped, or stopped -> running)
func (a Adapter) StartContainerStatusWatch() {

	go func() {

		// Map: key is container ID -> container state from Docker API
		previousStatus := map[string]*string{}

		for {

			containers, err := a.Client.GetContainerList(true)
			if err != nil {
				a.Logger().ReportError(errors.Wrap(err, "Error occurred on acquisition of container list"), machineoutput.TimestampNow())
			}

			if err == nil {

				containerStatusEntries := []machineoutput.ContainerStatusEntry{}

				componentContainers := a.Client.GetContainersByComponent(a.ComponentName, containers)

				// Locate containers which no longer exist
				for prevContainerID := range previousStatus {

					match := false
					for _, componentContainer := range componentContainers {
						if componentContainer.ID == prevContainerID {
							match = true
							break
						}
					}

					// Container no longer exists, so remove it from map and report it
					if !match {
						klog.V(4).Infof("Container %s does not exist, so reported as deleted.", prevContainerID)
						delete(previousStatus, prevContainerID)
						containerStatusEntries = append(containerStatusEntries, machineoutput.ContainerStatusEntry{
							ID:     prevContainerID,
							Status: "deleted",
						})
					}
				}

				for _, componentContainer := range componentContainers {

					klog.V(4).Infof("Container %s state was %s", componentContainer.ID, componentContainer.State)

					stateFromMap := previousStatus[componentContainer.ID]

					// If this is the first time we've seen this container, OR its state has changed
					if stateFromMap == nil || *stateFromMap != componentContainer.State {

						var componentContainerState string = componentContainer.State
						previousStatus[componentContainer.ID] = &componentContainerState

						containerStatusEntries = append(containerStatusEntries, machineoutput.ContainerStatusEntry{
							ID:     componentContainer.ID,
							Status: componentContainer.State,
						})
					}
				}

				// Log any change events
				if len(containerStatusEntries) > 0 {
					a.Logger().ContainerStatus(containerStatusEntries, machineoutput.TimestampNow())
				}

			}

			time.Sleep(ContainerCheckInterval)
		}
	}()
}

// StartSupervisordCtlStatusWatch kicks off a goroutine which calls 'supervisord ctl status' within every odo-managed container, every X seconds.
// If the status of the supervisord program changes (eg RUNNING <-> STOPPED), this change is reported to the console.
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
		// Map key: containerName (within pod) -> list of statuses from 'supervisord ctl status'
		lastContainerStatus := map[string][]supervisordStatus{}

		for {

			event := <-senderChannel

			previousStatus, hasLastContainerStatus := lastContainerStatus[event.containerName]
			lastContainerStatus[event.containerName] = event.status

			reportChange := false

			if hasLastContainerStatus {
				// If we saw a status from this container previously...

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

// querySupervisordStatusFromContainers runs 'supervisord ctl status' within each odo-managed container.
// The status results are sent to the reconciler.
func (sw *supervisordStatusWatcher) querySupervisordStatusFromContainers(a Adapter) {

	containers, err := utils.GetComponentContainers(a.Client, a.ComponentName)
	if err != nil {
		a.Logger().ReportError(errors.Wrap(err, "Unable to retrieve container status"), machineoutput.TimestampNow())
		return
	}

	// For each of the containers, retrieve the status of the programs and send that status back to the status reconciler
	for _, container := range containers {

		status := getSupervisordStatusInContainer(container.ID, a)

		sw.statusReconcilerChannel <- supervisordStatusEvent{
			containerName: container.ID,
			status:        status,
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

// getSupervisordStatusInContainer executes 'supervisord ctl status' within the container, parses the output,
// and returns the status
func getSupervisordStatusInContainer(containerID string, a Adapter) []supervisordStatus {

	command := []string{common.SupervisordBinaryPath, common.SupervisordCtlSubCommand, "status"}

	compInfo := common.ComponentInfo{
		ContainerName: containerID,
	}

	stdoutWriter, stdoutOutputChannel := common.CreateConsoleOutputWriterAndChannel()
	stderrWriter, stderrOutputChannel := common.CreateConsoleOutputWriterAndChannel()

	err := common.ExecuteCommand(&a.Client, compInfo, command, false, stdoutWriter, stderrWriter)

	// Close the writer and wait the console output
	stdoutWriter.Close()
	consoleResult := <-stdoutOutputChannel

	stderrWriter.Close()
	consoleStderrResult := <-stderrOutputChannel

	if err != nil {
		klog.V(4).Infof("Unable to execute command within container %s, %v, output: %v %v", containerID, err, consoleResult, consoleStderrResult)
		a.Logger().ReportError(err, machineoutput.TimestampNow())
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
	status        []supervisordStatus
}
