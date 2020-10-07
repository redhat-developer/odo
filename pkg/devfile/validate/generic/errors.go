package generic

import "fmt"

// InvalidEventError returns an error if the devfile event type has invalid events
type InvalidEventError struct {
	eventType string
	errorMsg  string
}

func (e *InvalidEventError) Error() string {
	return fmt.Sprintf("%s type events are invalid: %s", e.eventType, e.errorMsg)
}

// DuplicateCommandError returns an error if the command is duplicate
type DuplicateCommandError struct {
	commandId string
}

func (e *DuplicateCommandError) Error() string {
	return fmt.Sprintf("devfile has duplicate command IDs %q", e.commandId)
}

// ExecCommandMissingComponentError returns an error if the exec command does not have a component
type ExecCommandMissingComponentError struct {
	commandId string
}

func (e *ExecCommandMissingComponentError) Error() string {
	return fmt.Sprintf("exec command %q must reference a component", e.commandId)
}

// ExecCommandMissingCommandLineError returns an error if the exec command does not have a command line
type ExecCommandMissingCommandLineError struct {
	commandId string
}

func (e *ExecCommandMissingCommandLineError) Error() string {
	return fmt.Sprintf("exec command %q must have a command", e.commandId)
}

// InvalidCommandError returns an error if the command is an invalid type
type InvalidCommandError struct {
	commandId   string
	commandType string
}

func (e *InvalidCommandError) Error() string {
	return fmt.Sprintf("the command %q should be of type %s", e.commandId, e.commandType)
}

// ExecCommandInvalidContainerError returns an error if the exec command references an invalid container component
type ExecCommandInvalidContainerError struct {
	commandId string
}

func (e *ExecCommandInvalidContainerError) Error() string {
	return fmt.Sprintf("the command %q does not map to a container component", e.commandId)
}

// CompositeDirectReferenceError returns an error if the composite command directly references itself
type CompositeDirectReferenceError struct {
	commandId string
}

func (e *CompositeDirectReferenceError) Error() string {
	return fmt.Sprintf("the composite command %q cannot reference itself", e.commandId)
}

// CompositeIndirectReferenceError returns an error if the composite command indirectly references itself
type CompositeIndirectReferenceError struct {
	commandId string
}

func (e *CompositeIndirectReferenceError) Error() string {
	return fmt.Sprintf("the composite command %q cannot indirectly reference itself", e.commandId)
}

// CompositeMissingSubCommandError returns an error if the composite command has a missing sub command
type CompositeMissingSubCommandError struct {
	commandId  string
	subCommand string
}

func (e *CompositeMissingSubCommandError) Error() string {
	return fmt.Sprintf("the command %q mentioned in the composite command %q does not exist in the devfile", e.subCommand, e.commandId)
}

// CompositeInvalidSubCommandError returns an error if the composite command references an invalid sub command
type CompositeInvalidSubCommandError struct {
	commandId    string
	subCommandId string
	errorMsg     string
}

func (e *CompositeInvalidSubCommandError) Error() string {
	return fmt.Sprintf("the composite command %q references an invalid command %q: %s", e.commandId, e.subCommandId, e.errorMsg)
}

// ReservedEnvError returns an error if the user attempts to customize a reserved ENV in a container
type ReservedEnvError struct {
	componentName string
	envName       string
}

func (e *ReservedEnvError) Error() string {
	return fmt.Sprintf("env variable %s is reserved and cannot be customized in component %s", e.envName, e.componentName)
}

// InvalidVolumeSizeError returns an error if volume component has an invalid size
type InvalidVolumeSizeError struct {
	size            string
	componentName   string
	validationError error
}

func (e *InvalidVolumeSizeError) Error() string {
	return fmt.Sprintf("size %s for volume component %s is invalid: %v. Example - 2Gi, 1024Mi", e.size, e.componentName, e.validationError)
}

// DuplicateVolumeComponentsError returns an error if duplicate volume components are found
type DuplicateVolumeComponentsError struct {
}

func (e *DuplicateVolumeComponentsError) Error() string {
	return "duplicate volume components present in devfile"
}

// MissingVolumeMountError returns an error if the container volume mount does not reference a valid volume component
type MissingVolumeMountError struct {
	volumeName string
}

func (e *MissingVolumeMountError) Error() string {
	return fmt.Sprintf("unable to find volume mount %s in devfile volume components", e.volumeName)
}
