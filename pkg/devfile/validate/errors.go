package validate

import (
	"fmt"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// NoComponentsError returns an error if no component is found
type NoComponentsError struct {
}

func (e *NoComponentsError) Error() string {
	return "no components present"
}

// NoContainerComponentError returns an error if no container component is found
type NoContainerComponentError struct {
}

func (e *NoContainerComponentError) Error() string {
	return fmt.Sprintf("odo requires atleast one component of type '%s' in devfile", common.ContainerComponentType)
}

// DuplicateVolumeComponentsError returns an error if duplicate volume components are found
type DuplicateVolumeComponentsError struct {
}

func (e *DuplicateVolumeComponentsError) Error() string {
	return "duplicate volume components present in devfile"
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

// MissingVolumeMountError returns an error if the container volume mount does not reference a valid volume component
type MissingVolumeMountError struct {
	volumeName string
}

func (e *MissingVolumeMountError) Error() string {
	return fmt.Sprintf("unable to find volume mount %s in devfile volume components", e.volumeName)
}

// InvalidEventError returns an error if the devfile event type has invalid events
type InvalidEventError struct {
	eventType string
	errorMsg  string
}

func (e *InvalidEventError) Error() string {
	return fmt.Sprintf("%s type events is invalid: %s", e.eventType, e.errorMsg)
}

// UnsupportedOdoCommandError returns an error if the command is neither exec nor composite
type UnsupportedOdoCommandError struct {
	commandId string
}

func (e *UnsupportedOdoCommandError) Error() string {
	return fmt.Sprintf("command %q must be of type \"exec\" or \"composite\"", e.commandId)
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

// ExecCommandInvalidContainerError returns an error if the exec command references an invalid container component
type ExecCommandInvalidContainerError struct {
	commandId string
}

func (e *ExecCommandInvalidContainerError) Error() string {
	return fmt.Sprintf("the command %q does not map to a supported component", e.commandId)
}

// CompositeRunKindError returns an error if the composite command is of kind run
type CompositeRunKindError struct {
}

func (e *CompositeRunKindError) Error() string {
	return "composite commands of run Kind are not supported currently"
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
}

func (e *CompositeInvalidSubCommandError) Error() string {
	return fmt.Sprintf("the composite command %q references an invalid command %q", e.commandId, e.subCommandId)
}
