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

// InvalidCommandError returns an error if the command is invalid
type InvalidCommandError struct {
	commandId string
	reason    string
}

func (e *InvalidCommandError) Error() string {
	return fmt.Sprintf("the command %q is invalid - %s", e.commandId, e.reason)
}

// ReservedEnvError returns an error if the user attempts to customize a reserved ENV in a container
type ReservedEnvError struct {
	componentName string
	envName       string
}

func (e *ReservedEnvError) Error() string {
	return fmt.Sprintf("env variable %s is reserved and cannot be customized in component %s", e.envName, e.componentName)
}

// InvalidVolumeError returns an error if the volume is invalid
type InvalidVolumeError struct {
	name   string
	reason string
}

func (e *InvalidVolumeError) Error() string {
	return fmt.Sprintf("the volume %q is invalid - %s", e.name, e.reason)
}

// MissingVolumeMountError returns an error if the container volume mount does not reference a valid volume component
type MissingVolumeMountError struct {
	volumeName string
}

func (e *MissingVolumeMountError) Error() string {
	return fmt.Sprintf("unable to find volume mount %s in devfile volume components", e.volumeName)
}

// InvalidEndpointError returns an error if the component endpoint is invalid
type InvalidEndpointError struct {
	name string
	port int32
}

func (e *InvalidEndpointError) Error() string {
	var errMsg string
	if e.name != "" {
		errMsg = fmt.Sprintf("devfile contains multiple endpoint entries with same name: %v", e.name)
	} else if string(e.port) != "" {
		errMsg = fmt.Sprintf("devfile contains multiple containers with same TargetPort: %v", e.port)
	}

	return errMsg
}
