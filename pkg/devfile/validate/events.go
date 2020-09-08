package validate

import (
	"fmt"
	"strings"

	adapterCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/parser/data"
	"k8s.io/klog"
)

// validateEvents validates all the devfile events
func validateEvents(data data.DevfileData) error {

	eventErrors := ""
	events := data.GetEvents()

	switch {
	case len(events.PreStart) > 0:
		klog.V(2).Info("Validating preStart events")
		if preStartErr := isEventValid(data, events.PreStart, "preStart"); preStartErr != nil {
			eventErrors += fmt.Sprintf("\n%s", preStartErr.Error())
		}
		fallthrough
	case len(events.PostStart) > 0:
		klog.V(2).Info("Validating postStart events")
		if postStartErr := isEventValid(data, events.PostStart, "postStart"); postStartErr != nil {
			eventErrors += fmt.Sprintf("\n%s", postStartErr.Error())
		}
		fallthrough
	case len(events.PreStop) > 0:
		klog.V(2).Info("Validating preStop events")
		if preStopErr := isEventValid(data, events.PreStop, "preStop"); preStopErr != nil {
			eventErrors += fmt.Sprintf("\n%s", preStopErr.Error())
		}
		fallthrough
	case len(events.PostStop) > 0:
		klog.V(2).Info("Validating postStop events")
		if postStopErr := isEventValid(data, events.PostStop, "postStop"); postStopErr != nil {
			eventErrors += fmt.Sprintf("\n%s", postStopErr.Error())
		}
	}

	// if there is any validation error, return it
	if len(eventErrors) > 0 {
		return fmt.Errorf("devfile events validation error: %s", eventErrors)
	}

	return nil
}

// isEventValid checks if events belonging to a specific event type are valid:
// 1. Event should map to a valid devfile command
// 2. Event commands should be valid
func isEventValid(data data.DevfileData, eventNames []string, eventType string) error {
	eventErrorMsg := make(map[string][]string)
	eventErrors := ""
	commands := data.GetCommands()

	for _, eventName := range eventNames {
		isEventPresent := false
		for _, command := range commands {
			// Check if event matches a valid devfile command
			if command.GetID() == strings.ToLower(eventName) {
				isEventPresent = true

				// Check if the devfile command is valid
				err := adapterCommon.ValidateCommand(data, command)
				if err != nil {
					klog.V(2).Infof("command %s is not valid: %s", command.GetID(), err.Error())
					eventErrorMsg[strings.ToLower(eventName)] = append(eventErrorMsg[strings.ToLower(eventName)], err.Error())
				}
				break
			}
		}

		if !isEventPresent {
			klog.V(2).Infof("%s type event %s does not map to a valid devfile command", eventType, eventName)
			eventErrorMsg[strings.ToLower(eventName)] = append(eventErrorMsg[strings.ToLower(eventName)], fmt.Sprintf("event %s does not map to a valid devfile command", eventName))
		}
	}

	for eventName, errors := range eventErrorMsg {
		if len(errors) > 0 {
			klog.V(2).Infof("errors found for event %s belonging to %s: %s", eventName, eventType, strings.Join(errors, ","))
			eventErrors += fmt.Sprintf("\n%s is invalid: %s", eventName, strings.Join(errors, ","))
		}
	}

	if len(eventErrors) > 0 {
		klog.V(2).Infof("errors found for event type %s", eventType)
		return &InvalidEventError{eventType: eventType, errorMsg: eventErrors}
	}

	return nil
}
