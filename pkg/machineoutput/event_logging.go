package machineoutput

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"time"

	"k8s.io/klog"
)

// formatTime returns time in UTC Unix Epoch Seconds and then the microsecond portion of that time.
func formatTime(time time.Time) string {
	result := fmt.Sprintf("%d.%06d", time.Unix(), time.Nanosecond()/1000)
	return result

}

// TimestampNow returns timestamp in format of (seconds since UTC Unix epoch).(microseconds time component)
func TimestampNow() string {
	return formatTime(time.Now())
}

// NewNoOpMachineEventLoggingClient creates a new instance of NoOpMachineEventLoggingClient,
// which will ignore any provided events.
func NewNoOpMachineEventLoggingClient() *NoOpMachineEventLoggingClient {
	return &NoOpMachineEventLoggingClient{}
}

// DevFileCommandExecutionBegin ignores the provided event.
func (c *NoOpMachineEventLoggingClient) DevFileCommandExecutionBegin(commandName string, timestamp string) {
}

// DevFileCommandExecutionComplete ignores the provided event.
func (c *NoOpMachineEventLoggingClient) DevFileCommandExecutionComplete(commandName string, timestamp string, errorVal error) {
}

// CreateContainerOutputWriter ignores the provided event.
func (c *NoOpMachineEventLoggingClient) CreateContainerOutputWriter(stderr bool) io.Writer {
	return nil
}

// ReportError ignores the provided event.
func (c *NoOpMachineEventLoggingClient) ReportError(errorVal error, timestamp string) {}

// NewConsoleMachineEventLoggingClient creates a new instance of ConsoleMachineEventLoggingClient,
// which will output events as JSON to the console.
func NewConsoleMachineEventLoggingClient() *ConsoleMachineEventLoggingClient {
	return &ConsoleMachineEventLoggingClient{}
}

// DevFileCommandExecutionBegin outputs the provided event as JSON to the console.
func (c *ConsoleMachineEventLoggingClient) DevFileCommandExecutionBegin(commandName string, timestamp string) {

	json := MachineEventWrapper{
		DevFileCommandExecutionBegin: &DevFileCommandExecutionBegin{
			CommandName:      commandName,
			AbstractLogEvent: AbstractLogEvent{Timestamp: timestamp},
		},
	}

	OutputSuccessUnindented(json)
}

// DevFileCommandExecutionComplete outputs the provided event as JSON to the console.
func (c *ConsoleMachineEventLoggingClient) DevFileCommandExecutionComplete(commandName string, timestamp string, errorVal error) {

	var errorStr string

	if errorVal != nil {
		errorStr = errorVal.Error()
	}

	json := MachineEventWrapper{
		DevFileCommandExecutionComplete: &DevFileCommandExecutionComplete{
			CommandName:      commandName,
			AbstractLogEvent: AbstractLogEvent{Timestamp: timestamp},
			Error:            errorStr,
		},
	}

	OutputSuccessUnindented(json)
}

// CreateContainerOutputWriter returns an io.Writer for which the devfile command/action process output should be
// written (for example by passing the io.Writer to exec.ExecuteCommand).
//
// All text written to the returned object will be output as a log text event.
func (c *ConsoleMachineEventLoggingClient) CreateContainerOutputWriter(stderr bool) io.Writer {
	reader, writer := io.Pipe()

	stream := "stdout"
	if stderr {
		stream = "stderr"
	}

	go func() {

		bufReader := bufio.NewReader(reader)
		for {
			line, _, err := bufReader.ReadLine()
			if err != nil {
				klog.V(4).Infof("Unexpected error on reading container output reader: %v", err)
				return
			}

			json := MachineEventWrapper{
				LogText: &LogText{
					AbstractLogEvent: AbstractLogEvent{Timestamp: TimestampNow()},
					Text:             string(line),
					Stream:           stream,
				},
			}
			OutputSuccessUnindented(json)
		}

	}()

	return writer
}

// ReportError ignores the provided event.
func (c *ConsoleMachineEventLoggingClient) ReportError(errorVal error, timestamp string) {
	json := MachineEventWrapper{
		ReportError: &ReportError{
			Error:            errorVal.Error(),
			AbstractLogEvent: AbstractLogEvent{Timestamp: timestamp},
		},
	}

	OutputSuccessUnindented(json)
}

// GetEntry will return the JSON event parsed from a single line of '-o json' machine readable console output.
func (w MachineEventWrapper) GetEntry() (MachineEventLogEntry, error) {

	if w.DevFileCommandExecutionBegin != nil {
		return w.DevFileCommandExecutionBegin, nil

	} else if w.DevFileCommandExecutionComplete != nil {
		return w.DevFileCommandExecutionComplete, nil

	} else if w.LogText != nil {
		return w.LogText, nil

	} else if w.ReportError != nil {
		return w.ReportError, nil

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
)

// GetCommandName returns a command if the MLE supports that field (otherwise empty string is returned).
// Currently used for test purposes only.
func GetCommandName(entry MachineEventLogEntry) string {

	if entry.GetType() == TypeDevFileCommandExecutionBegin {
		return entry.(*DevFileCommandExecutionBegin).CommandName
	} else if entry.GetType() == TypeDevFileCommandExecutionComplete {
		return entry.(*DevFileCommandExecutionComplete).CommandName
	} else {
		return ""
	}

}

// FindNextEntryByType locates the next entry of a given type within a slice. Currently used for test purposes only.
func FindNextEntryByType(initialIndex int, typeToFind MachineEventLogEntryType, entries []MachineEventLogEntry) (MachineEventLogEntry, int) {

	for index, entry := range entries {
		if index < initialIndex {
			continue
		}

		if entry.GetType() == typeToFind {
			return entry, index
		}
	}

	return nil, -1

}
