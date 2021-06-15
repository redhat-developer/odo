package validation

import (
	"fmt"
	attributesAPI "github.com/devfile/api/v2/pkg/attributes"
)

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

// InvalidCommandError returns an error if the command is invalid
type InvalidCommandTypeError struct {
	commandId string
}

func (e *InvalidCommandTypeError) Error() string {
	return fmt.Sprintf("command %s has invalid type", e.commandId)
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
	errMsg string
}

func (e *MissingVolumeMountError) Error() string {
	return fmt.Sprintf("unable to find the following volume mounts in devfile volume components: %s", e.errMsg)
}

// InvalidEndpointError returns an error if the component endpoint is invalid
type InvalidEndpointError struct {
	name string
	port int
}

func (e *InvalidEndpointError) Error() string {
	var errMsg string
	if e.name != "" {
		errMsg = fmt.Sprintf("devfile contains multiple endpoint entries with same name: %v", e.name)
	} else if fmt.Sprint(e.port) != "" {
		errMsg = fmt.Sprintf("devfile contains multiple containers with same TargetPort: %v", e.port)
	}

	return errMsg
}

// InvalidComponentError returns an error if the component is invalid
type InvalidComponentError struct {
	componentName string
	reason        string
}

func (e *InvalidComponentError) Error() string {
	return fmt.Sprintf("the component %q is invalid - %s", e.componentName, e.reason)
}

// resolveErrorMessageWithImportAttributes returns an updated error message
// with detailed information on the imported and overriden resource.
// example:
// "the component <compName> is invalid - <reason>, imported from Uri: http://example.com/devfile.yaml, in parent overrides from main devfile"
func resolveErrorMessageWithImportAttributes(validationErr error, attributes attributesAPI.Attributes) error {
	var findKeyErr error
	importReference := attributes.Get(ImportSourceAttribute, &findKeyErr)

	// overridden element must contain import resource information
	// an overridden element can be either parentOverride or pluginOverride
	// example:
	// if an element is imported from another devfile, but contains no overrides - ImportSourceAttribute
	// if an element is from parentOverride - ImportSourceAttribute + ParentOverrideAttribute
	// if an element is from pluginOverride - ImportSourceAttribute + PluginOverrideAttribute
	if findKeyErr == nil {
		validationErr = fmt.Errorf("%s, imported from %s", validationErr.Error(), importReference)
		parentOverrideReference := attributes.Get(ParentOverrideAttribute, &findKeyErr)
		if findKeyErr == nil {
			validationErr = fmt.Errorf("%s, in parent overrides from %s", validationErr.Error(), parentOverrideReference)
		} else {
			// reset findKeyErr to nil
			findKeyErr = nil
			pluginOverrideReference := attributes.Get(PluginOverrideAttribute, &findKeyErr)
			if findKeyErr == nil {
				validationErr = fmt.Errorf("%s, in plugin overrides from %s", validationErr.Error(), pluginOverrideReference)
			}
		}
	}

	return validationErr
}
