package machineoutput

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/openshift/odo/v2/pkg/log"

	"k8s.io/klog"
)

// FormatTime returns time in UTC Unix Epoch Seconds and then the microsecond portion of that time.
func FormatTime(time time.Time) string {
	result := fmt.Sprintf("%d.%06d", time.Unix(), time.Nanosecond()/1000)
	return result

}

// TimestampNow returns timestamp in format of (seconds since UTC Unix epoch).(microseconds time component)
func TimestampNow() string {
	return FormatTime(time.Now())
}

// NewMachineEventLoggingClient creates the appropriate client based on whether we are in machine logging mode or not
func NewMachineEventLoggingClient() MachineEventLoggingClient {
	if log.IsJSON() {
		return NewConsoleMachineEventLoggingClient()
	}

	return NewNoOpMachineEventLoggingClient()
}

// NewNoOpMachineEventLoggingClient creates a new instance of NoOpMachineEventLoggingClient,
// which will ignore any provided events.
func NewNoOpMachineEventLoggingClient() *NoOpMachineEventLoggingClient {
	return &NoOpMachineEventLoggingClient{}
}

var _ MachineEventLoggingClient = &NoOpMachineEventLoggingClient{}

// DevFileCommandExecutionBegin ignores the provided event.
func (c *NoOpMachineEventLoggingClient) DevFileCommandExecutionBegin(commandID string, componentName string, commandLine string, groupKind string, timestamp string) {
}

// DevFileCommandExecutionComplete ignores the provided event.
func (c *NoOpMachineEventLoggingClient) DevFileCommandExecutionComplete(commandID string, componentName string, commandLine string, groupKind string, timestamp string, errorVal error) {
}

// CreateContainerOutputWriter ignores the provided event.
func (c *NoOpMachineEventLoggingClient) CreateContainerOutputWriter() (*io.PipeWriter, chan interface{}, *io.PipeWriter, chan interface{}) {

	channels := []chan interface{}{make(chan interface{}), make(chan interface{})}

	// Ensure there is always a result waiting on each of the channels
	for _, channelPtr := range channels {
		channelVal := channelPtr

		go func(channel chan interface{}) {
			for {
				channel <- nil
			}
		}(channelVal)
	}

	return nil, channels[0], nil, channels[1]
}

// ReportError ignores the provided event.
func (c *NoOpMachineEventLoggingClient) ReportError(errorVal error, timestamp string) {}

// SupervisordStatus ignores the provided event.
func (c *NoOpMachineEventLoggingClient) SupervisordStatus(statuses []SupervisordStatusEntry, timestamp string) {
}

// ContainerStatus ignores the provided event.
func (c *NoOpMachineEventLoggingClient) ContainerStatus(statuses []ContainerStatusEntry, timestamp string) {
}

// URLReachable ignores the provided event.
func (c *NoOpMachineEventLoggingClient) URLReachable(name string, url string, port int, secure bool, kind string, reachable bool, timestamp string) {

}

// KubernetesPodStatus ignores the provided event.
func (c *NoOpMachineEventLoggingClient) KubernetesPodStatus(pods []KubernetesPodStatusEntry, timestamp string) {

}

// NewConsoleMachineEventLoggingClient creates a new instance of ConsoleMachineEventLoggingClient,
// which will output events as JSON to the console.
func NewConsoleMachineEventLoggingClient() *ConsoleMachineEventLoggingClient {
	return &ConsoleMachineEventLoggingClient{}
}

// NewConsoleMachineEventLoggingClientWithFunction creates a new instance with a custom logging functional,
// suitable for, for example, unit testing.
func NewConsoleMachineEventLoggingClientWithFunction(logFunc func(machineOutput MachineEventWrapper)) *ConsoleMachineEventLoggingClient {
	return &ConsoleMachineEventLoggingClient{
		logFunc: logFunc,
	}
}

var _ MachineEventLoggingClient = &ConsoleMachineEventLoggingClient{}

// DevFileCommandExecutionBegin outputs the provided event as JSON to the console.
func (c *ConsoleMachineEventLoggingClient) DevFileCommandExecutionBegin(commandID string, componentName string, commandLine string, groupKind string, timestamp string) {

	json := MachineEventWrapper{
		DevFileCommandExecutionBegin: &DevFileCommandExecutionBegin{
			CommandID:        commandID,
			ComponentName:    componentName,
			CommandLine:      commandLine,
			GroupKind:        groupKind,
			AbstractLogEvent: AbstractLogEvent{Timestamp: timestamp},
		},
	}

	c.outputJSON(json)
}

// DevFileCommandExecutionComplete outputs the provided event as JSON to the console.
func (c *ConsoleMachineEventLoggingClient) DevFileCommandExecutionComplete(commandID string, componentName string, commandLine string, groupKind string, timestamp string, errorVal error) {

	errorStr := ""

	if errorVal != nil {
		errorStr = errorVal.Error()
	}

	json := MachineEventWrapper{
		DevFileCommandExecutionComplete: &DevFileCommandExecutionComplete{
			DevFileCommandExecutionBegin: DevFileCommandExecutionBegin{
				CommandID:        commandID,
				ComponentName:    componentName,
				CommandLine:      commandLine,
				GroupKind:        groupKind,
				AbstractLogEvent: AbstractLogEvent{Timestamp: timestamp},
			},
			Error: errorStr,
		},
	}

	c.outputJSON(json)
}

// CreateContainerOutputWriter returns an io.PipeWriter for which the devfile command/action process output should be
// written (for example by passing the io.PipeWriter to exec.ExecuteCommand), and a channel for communicating when the last data
// has been received on the reader.
//
// All text written to the returned object will be output as a log text event.
// Returned channels will each contain a single nil entry once the underlying reader has closed.
func (c *ConsoleMachineEventLoggingClient) CreateContainerOutputWriter() (*io.PipeWriter, chan interface{}, *io.PipeWriter, chan interface{}) {

	stdoutWriter, stdoutChannel := createWriterAndChannel(false)
	stderrWriter, stderrChannel := createWriterAndChannel(true)

	return stdoutWriter, stdoutChannel, stderrWriter, stderrChannel

}

// ReportError outputs the provided event as JSON to the console.
func (c *ConsoleMachineEventLoggingClient) ReportError(errorVal error, timestamp string) {
	json := MachineEventWrapper{
		ReportError: &ReportError{
			Error:            errorVal.Error(),
			AbstractLogEvent: AbstractLogEvent{Timestamp: timestamp},
		},
	}

	c.outputJSON(json)
}

// SupervisordStatus outputs the provided event as JSON to the console.
func (c *ConsoleMachineEventLoggingClient) SupervisordStatus(statuses []SupervisordStatusEntry, timestamp string) {
	json := MachineEventWrapper{
		SupervisordStatus: &SupervisordStatus{
			ProgramStatus:    statuses,
			AbstractLogEvent: AbstractLogEvent{Timestamp: timestamp},
		},
	}

	c.outputJSON(json)
}

// ContainerStatus outputs the provided event as JSON to the console.
func (c *ConsoleMachineEventLoggingClient) ContainerStatus(statuses []ContainerStatusEntry, timestamp string) {
	json := MachineEventWrapper{
		ContainerStatus: &ContainerStatus{
			Status: statuses,
			AbstractLogEvent: AbstractLogEvent{
				Timestamp: timestamp,
			},
		},
	}
	c.outputJSON(json)
}

// URLReachable outputs the provided event as JSON to the console.
func (c *ConsoleMachineEventLoggingClient) URLReachable(name string, url string, port int, secure bool, kind string, reachable bool, timestamp string) {
	json := MachineEventWrapper{
		URLReachable: &URLReachable{
			Name:             name,
			URL:              url,
			Port:             port,
			Secure:           secure,
			Kind:             kind,
			Reachable:        reachable,
			AbstractLogEvent: AbstractLogEvent{Timestamp: timestamp},
		},
	}
	c.outputJSON(json)
}

// KubernetesPodStatus outputs the provided event as JSON to the console.
func (c *ConsoleMachineEventLoggingClient) KubernetesPodStatus(pods []KubernetesPodStatusEntry, timestamp string) {
	json := MachineEventWrapper{
		KubernetesPodStatus: &KubernetesPodStatus{
			Pods: pods,
			AbstractLogEvent: AbstractLogEvent{
				Timestamp: timestamp,
			},
		},
	}
	c.outputJSON(json)
}

func (c *ConsoleMachineEventLoggingClient) outputJSON(machineOutput MachineEventWrapper) {

	if c.logFunc != nil {
		c.logFunc(machineOutput)
		return
	}

	OutputSuccessUnindented(machineOutput)
}

// GetEntry will return the JSON event parsed from a single line of '-o json' machine readable console output.
// Currently used for test purposes only.
func (w MachineEventWrapper) GetEntry() (MachineEventLogEntry, error) {

	if w.DevFileCommandExecutionBegin != nil {
		return w.DevFileCommandExecutionBegin, nil

	} else if w.DevFileCommandExecutionComplete != nil {
		return w.DevFileCommandExecutionComplete, nil

	} else if w.LogText != nil {
		return w.LogText, nil

	} else if w.ReportError != nil {
		return w.ReportError, nil

	} else if w.KubernetesPodStatus != nil {
		return w.KubernetesPodStatus, nil

	} else if w.ContainerStatus != nil {
		return w.ContainerStatus, nil

	} else if w.SupervisordStatus != nil {
		return w.SupervisordStatus, nil

	} else if w.URLReachable != nil {
		return w.URLReachable, nil

	} else {
		return nil, errors.New("unexpected machine event log entry")
	}

}

// GetTimestamp returns the timestamp element for this event.
func (c AbstractLogEvent) GetTimestamp() string { return c.Timestamp }

// GetType returns the event type for this event.
func (c DevFileCommandExecutionBegin) GetType() MachineEventLogEntryType {
	return TypeDevFileCommandExecutionBegin
}

// GetType returns the event type for this event.
func (c DevFileCommandExecutionComplete) GetType() MachineEventLogEntryType {
	return TypeDevFileCommandExecutionComplete
}

// GetType returns the event type for this event.
func (c LogText) GetType() MachineEventLogEntryType { return TypeLogText }

// GetType returns the event type for this event.
func (c ReportError) GetType() MachineEventLogEntryType { return TypeReportError }

// GetType returns the event type for this event.
func (c SupervisordStatus) GetType() MachineEventLogEntryType { return TypeSupervisordStatus }

// GetType returns the event type for this event.
func (c ContainerStatus) GetType() MachineEventLogEntryType { return TypeContainerStatus }

// GetType returns the event type for this event.
func (c URLReachable) GetType() MachineEventLogEntryType { return TypeURLReachable }

// GetType returns the event type for this event.
func (c KubernetesPodStatus) GetType() MachineEventLogEntryType { return TypeKubernetesPodStatus }

// MachineEventLogEntryType indicates the machine-readable event type from an ODO operation
type MachineEventLogEntryType int

const (
	// TypeDevFileCommandExecutionBegin is the entry type for that event.
	TypeDevFileCommandExecutionBegin MachineEventLogEntryType = 0
	// TypeDevFileCommandExecutionComplete is the entry type for that event.
	TypeDevFileCommandExecutionComplete MachineEventLogEntryType = 1
	// TypeLogText is the entry type for that event.
	TypeLogText MachineEventLogEntryType = 2
	// TypeReportError is the entry type for that event.
	TypeReportError MachineEventLogEntryType = 3
	// TypeSupervisordStatus is the entry type for that event.
	TypeSupervisordStatus MachineEventLogEntryType = 4
	// TypeContainerStatus is the entry type for that event.
	TypeContainerStatus MachineEventLogEntryType = 5
	// TypeURLReachable is the entry type for that event.
	TypeURLReachable MachineEventLogEntryType = 6
	// TypeKubernetesPodStatus is the entry type for that event.
	TypeKubernetesPodStatus MachineEventLogEntryType = 7
)

// GetCommandName returns a command if the MLE supports that field (otherwise empty string is returned).
// Currently used for test purposes only.
func GetCommandName(entry MachineEventLogEntry) string {

	if entry.GetType() == TypeDevFileCommandExecutionBegin {
		return entry.(*DevFileCommandExecutionBegin).CommandID
	} else if entry.GetType() == TypeDevFileCommandExecutionComplete {
		return entry.(*DevFileCommandExecutionComplete).CommandID
	} else {
		return ""
	}

}

// createWriterAndChannel is similar to the exec.CreateConsoleOutputWriterAndChannel(); see that function's comment for details.
func createWriterAndChannel(stderr bool) (*io.PipeWriter, chan interface{}) {
	reader, writer := io.Pipe()

	closeChannel := make(chan interface{})

	stream := "stdout"
	if stderr {
		stream = "stderr"
	}

	go func() {

		bufReader := bufio.NewReader(reader)
		for {
			line, _, err := bufReader.ReadLine()
			if err != nil {
				if err != io.EOF {
					klog.V(4).Infof("Unexpected error on reading container output reader: %v", err)
				}
				break
			}

			// Output log text event for each line we receive
			json := MachineEventWrapper{
				LogText: &LogText{
					AbstractLogEvent: AbstractLogEvent{Timestamp: TimestampNow()},
					Text:             string(line),
					Stream:           stream,
				},
			}
			OutputSuccessUnindented(json)
		}

		// Output a single nil event on the channel to inform that the last line of text has been
		// received from the writer.
		closeChannel <- nil
	}()

	return writer, closeChannel
}
