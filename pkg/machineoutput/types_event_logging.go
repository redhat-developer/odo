package machineoutput

import (
	"io"
)

// MachineEventLoggingClient is an interface which is used by consuming code to output machine-readable
// event JSON to the console. Both no-op and non-no-op implementations of this interface exist.
type MachineEventLoggingClient interface {

	// These functions output the corresponding eponymous JSON event to the console

	DevFileCommandExecutionBegin(commandID string, componentName string, commandLine string, groupKind string, timestamp string)
	DevFileCommandExecutionComplete(commandID string, componentName string, commandLine string, groupKind string, timestamp string, errorVal error)
	ReportError(errorVal error, timestamp string)

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
}

// DevFileCommandExecutionBegin is the JSON event that is emitted when a dev file command begins execution.
type DevFileCommandExecutionBegin struct {
	CommandID     string `json:"commandId"`
	ComponentName string `json:"componentName"`
	CommandLine   string `json:"commandLine"`
	GroupKind     string `json:"groupKind"`
	Timestamp     string `json:"timestamp"`
	AbstractLogEvent
}

// DevFileCommandExecutionComplete is the JSON event that is emitted when a dev file command completes execution.
type DevFileCommandExecutionComplete struct {
	CommandID     string `json:"commandId"`
	ComponentName string `json:"componentName"`
	CommandLine   string `json:"commandLine"`
	GroupKind     string `json:"groupKind"`
	Timestamp     string `json:"timestamp"`
	Error         string `json:"error,omitempty"`
	AbstractLogEvent
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

// AbstractLogEvent is the base struct for all events; all events must at a minimum contain a timestamp.
type AbstractLogEvent struct {
	Timestamp string `json:"timestamp"`
}

// Ensure the various events correctly implement the desired interface.
var _ MachineEventLogEntry = &DevFileCommandExecutionBegin{}
var _ MachineEventLogEntry = &DevFileCommandExecutionComplete{}
var _ MachineEventLogEntry = &LogText{}
var _ MachineEventLogEntry = &ReportError{}

// MachineEventLogEntry contains the expected methods for every event that is emitted.
// (This is mainly used for test purposes.)
type MachineEventLogEntry interface {
	GetTimestamp() string
	GetType() MachineEventLogEntryType
}

// NoOpMachineEventLoggingClient will ignore (eg not output) all events passed to it
type NoOpMachineEventLoggingClient struct {
}

// ConsoleMachineEventLoggingClient will output all events to the console as JSON
type ConsoleMachineEventLoggingClient struct {
}
