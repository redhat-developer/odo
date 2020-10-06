package validate

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	genericValidation "github.com/openshift/odo/pkg/devfile/validate/generic"
	"k8s.io/klog"
)

// validateEvents validates all the devfile events
func validateEvents(events common.DevfileEvents, commands map[string]common.DevfileCommand, components []common.DevfileComponent) error {

	eventErrors := ""
	var preStartGenericErr, postStartGenericErr, preStopGenericErr, postStopGenericErr error

	switch {
	case len(events.PreStart) > 0:
		klog.V(2).Info("Validating preStart events")
		if preStartGenericErr = genericValidation.IsEventValid(events.PreStart, "preStart", commands); preStartGenericErr != nil {
			eventErrors += fmt.Sprintf("\n%s", preStartGenericErr.Error())
		}
		if preStartErr := isEventValid(events.PreStart, "preStart", commands, components); preStartGenericErr == nil && preStartErr != nil {
			eventErrors += fmt.Sprintf("\n%s", preStartErr.Error())
		}
		fallthrough
	case len(events.PostStart) > 0:
		klog.V(2).Info("Validating postStart events")
		if postStartGenericErr = genericValidation.IsEventValid(events.PostStart, "postStart", commands); postStartGenericErr != nil {
			eventErrors += fmt.Sprintf("\n%s", postStartGenericErr.Error())
		}
		if postStartErr := isEventValid(events.PostStart, "postStart", commands, components); postStartGenericErr == nil && postStartErr != nil {
			eventErrors += fmt.Sprintf("\n%s", postStartErr.Error())
		}
		fallthrough
	case len(events.PreStop) > 0:
		klog.V(2).Info("Validating preStop events")
		if preStopGenericErr = genericValidation.IsEventValid(events.PreStop, "preStop", commands); preStopGenericErr != nil {
			eventErrors += fmt.Sprintf("\n%s", preStopGenericErr.Error())
		}
		if preStopErr := isEventValid(events.PreStop, "preStop", commands, components); preStopGenericErr == nil && preStopErr != nil {
			eventErrors += fmt.Sprintf("\n%s", preStopErr.Error())
		}
		fallthrough
	case len(events.PostStop) > 0:
		klog.V(2).Info("Validating postStop events")
		if postStopGenericErr = genericValidation.IsEventValid(events.PostStop, "postStop", commands); postStopGenericErr != nil {
			eventErrors += fmt.Sprintf("\n%s", postStopGenericErr.Error())
		}
		if postStopErr := isEventValid(events.PostStop, "postStop", commands, components); postStopGenericErr == nil && postStopErr != nil {
			eventErrors += fmt.Sprintf("\n%s", postStopErr.Error())
		}
	}

	// if there is any validation error, return it
	if len(eventErrors) > 0 {
		return fmt.Errorf("devfile events validation error: %s", eventErrors)
	}

	return nil
}

// isEventValid checks if events belonging to a specific event type are valid ie; either exec or composite command
func isEventValid(eventNames []string, eventType string, commands map[string]common.DevfileCommand, components []common.DevfileComponent) error {
	eventErrorMsg := make(map[string][]string)
	eventErrors := ""

	for _, eventName := range eventNames {
		for _, command := range commands {
			if command.GetID() == strings.ToLower(eventName) {
				// Check if the devfile command is a valid odo devfile command
				err := validateCommand(command, commands, components)
				if err != nil {
					klog.V(2).Infof("command %s is not valid: %s", command.GetID(), err.Error())
					eventErrorMsg[strings.ToLower(eventName)] = append(eventErrorMsg[strings.ToLower(eventName)], err.Error())
				}
				break
			}
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
