package validate

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"k8s.io/klog"
)

// validateEvents validates all the devfile events
func validateEvents(events common.DevfileEvents, commands map[string]common.DevfileCommand, components []common.DevfileComponent) error {

	eventErrors := ""

	switch {
	case len(events.PreStart) > 0:
		klog.V(2).Info("Validating preStart events")
		if preStartErr := isEventValid(events.PreStart, "preStart", commands, components); preStartErr != nil {
			eventErrors += fmt.Sprintf("\n%s", preStartErr.Error())
		}
		fallthrough
	case len(events.PostStart) > 0:
		klog.V(2).Info("Validating postStart events")
		if postStartErr := isEventValid(events.PostStart, "postStart", commands, components); postStartErr != nil {
			eventErrors += fmt.Sprintf("\n%s", postStartErr.Error())
		}
		fallthrough
	case len(events.PreStop) > 0:
		klog.V(2).Info("Validating preStop events")
		if preStopErr := isEventValid(events.PreStop, "preStop", commands, components); preStopErr != nil {
			eventErrors += fmt.Sprintf("\n%s", preStopErr.Error())
		}
		fallthrough
	case len(events.PostStop) > 0:
		klog.V(2).Info("Validating postStop events")
		if postStopErr := isEventValid(events.PostStop, "postStop", commands, components); postStopErr != nil {
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
func isEventValid(eventNames []string, eventType string, commands map[string]common.DevfileCommand, components []common.DevfileComponent) error {
	eventErrorMsg := make(map[string][]string)
	eventErrors := ""

	for _, eventName := range eventNames {
		isEventPresent := false
		for _, command := range commands {
			// Check if event matches a valid devfile command
			if command.GetID() == strings.ToLower(eventName) {
				isEventPresent = true

				// Check if the devfile command is valid
				err := validateCommand(command, commands, components)
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
