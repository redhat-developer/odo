package machineoutput

import (
	"io"

	corev1 "k8s.io/api/core/v1"
)

// MachineEventLoggingClient is an interface which is used by consuming code to output machine-readable
// event JSON to the console. Both no-op and non-no-op implementations of this interface exist.
type MachineEventLoggingClient interface {

	// These functions output the corresponding eponymous JSON event to the console

	DevFileCommandExecutionBegin(commandID string, componentName string, commandLine string, groupKind string, timestamp string)
	DevFileCommandExecutionComplete(commandID string, componentName string, commandLine string, groupKind string, timestamp string, errorVal error)
	ReportError(errorVal error, timestamp string)

	ContainerStatus(statuses []ContainerStatusEntry, timestamp string)

	URLReachable(name string, url string, port int, secure bool, kind string, reachable bool, timestamp string)

	KubernetesPodStatus(pods []KubernetesPodStatusEntry, timestamp string)

	// CreateContainerOutputWriter is used to capture output from container processes, and synchronously write it to the screen as LogText. See implementation comments for details.
	CreateContainerOutputWriter() (*io.PipeWriter, chan interface{}, *io.PipeWriter, chan interface{})
}

// MachineEventWrapper - a single line of machine-readable event console output must contain only one
// of these commands; the MachineEventWrapper is used to create (and parse, for tests) these lines.
type MachineEventWrapper struct {
	DevFileCommandExecutionBegin    *DevFileCommandExecutionBegin    `json:"devFileCommandExecutionBegin,omitempty"`
	DevFileCommandExecutionComplete *DevFileCommandExecutionComplete `json:"devFileCommandExecutionComplete,omitempty"`
	LogText                         *LogText                         `json:"logText,omitempty"`
	ReportError                     *ReportError                     `json:"reportError,omitempty"`
	ContainerStatus                 *ContainerStatus                 `json:"containerStatus,omitempty"`
	URLReachable                    *URLReachable                    `json:"urlReachable,omitempty"`
	KubernetesPodStatus             *KubernetesPodStatus             `json:"kubernetesPodStatus,omitempty"`
}

// DevFileCommandExecutionBegin is the JSON event that is emitted when a dev file command begins execution.
type DevFileCommandExecutionBegin struct {
	CommandID     string `json:"commandId"`
	ComponentName string `json:"componentName"`
	CommandLine   string `json:"commandLine"`
	GroupKind     string `json:"groupKind"`
	AbstractLogEvent
}

// DevFileCommandExecutionComplete is the JSON event that is emitted when a dev file command completes execution.
type DevFileCommandExecutionComplete struct {
	DevFileCommandExecutionBegin
	Error string `json:"error,omitempty"`
}

// ReportError is the JSON event that is emitted when an error occurs during push command
type ReportError struct {
	Error string `json:"error"`
	AbstractLogEvent
}

// LogText is the JSON event that is emitted when a dev file action outputs text to the console.
type LogText struct {
	Text   string `json:"text"`
	Stream string `json:"stream"`
	AbstractLogEvent
}

// ContainerStatus is the JSON event that is emitted to indicate odo-managed Docker container status
type ContainerStatus struct {
	Status []ContainerStatusEntry `json:"status"`
	AbstractLogEvent
}

// ContainerStatusEntry is an individual container's status
type ContainerStatusEntry struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// URLReachable is the JSON event that is emitted to indicate whether one of the component's URL's could be reached.
type URLReachable struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	Port      int    `json:"port"`
	Secure    bool   `json:"secure"`
	Kind      string `json:"kind"`
	Reachable bool   `json:"reachable"`
	AbstractLogEvent
}

// KubernetesPodStatus is the JSON event that emitted to indicate the status of pods in an odo-managed deployment
type KubernetesPodStatus struct {
	Pods []KubernetesPodStatusEntry `json:"pods"`
	AbstractLogEvent
}

// KubernetesPodStatusEntry is an individual pod's information
type KubernetesPodStatusEntry struct {
	Name           string                   `json:"name"`
	UID            string                   `json:"uid"`
	Phase          string                   `json:"phase"`
	Labels         map[string]string        `json:"labels,omitempty"`
	StartTime      string                   `json:"startTime,omitempty"`
	Containers     []corev1.ContainerStatus `json:"containers"`
	InitContainers []corev1.ContainerStatus `json:"initContainers"`
	// This embeds the K8s ContainerStatus API into the log output; further experimentation is required by
	// consuming tools to determine which fields from this struct are useful/reliable, at which point this should
	// be replaceable with a pared-down version containing only those fields. My early analysis is that the
	// vast majority are useful.
}

// AbstractLogEvent is the base struct for all events; all events must at a minimum contain a timestamp.
type AbstractLogEvent struct {
	Timestamp string `json:"timestamp"`
}

// Ensure the various events correctly implement the desired interface.
var _ MachineEventLogEntry = &DevFileCommandExecutionBegin{}
var _ MachineEventLogEntry = &DevFileCommandExecutionComplete{}
var _ MachineEventLogEntry = &LogText{}
var _ MachineEventLogEntry = &ReportError{}
var _ MachineEventLogEntry = &ContainerStatus{}
var _ MachineEventLogEntry = &URLReachable{}
var _ MachineEventLogEntry = &KubernetesPodStatus{}

// MachineEventLogEntry contains the expected methods for every event that is emitted.
// (This is mainly used for test purposes.)
type MachineEventLogEntry interface {
	GetTimestamp() string
}

// NoOpMachineEventLoggingClient will ignore (eg not output) all events passed to it
type NoOpMachineEventLoggingClient struct {
}

var _ MachineEventLoggingClient = (*NoOpMachineEventLoggingClient)(nil)

// ConsoleMachineEventLoggingClient will output all events to the console as JSON
type ConsoleMachineEventLoggingClient struct {

	// logFunc is an optional function that can be used instead of writing via the standard machine out logic
	logFunc func(machineOutput MachineEventWrapper)
}

var _ MachineEventLoggingClient = (*ConsoleMachineEventLoggingClient)(nil)
