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
	return fmt.Sprintf("the command %q does not map to a container component", e.commandId)
}

// CompositeRunKindError returns an error if the composite command is of kind run
type CompositeRunKindError struct {
}

func (e *CompositeRunKindError) Error() string {
	return "composite commands of run Kind are not supported currently"
}
