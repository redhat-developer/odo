package generic

import (
	"fmt"
	"strings"

	devfilev1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	"k8s.io/klog"
)

// validateEvents validates all the devfile events
func validateEvents(events devfilev1.Events, commands map[string]devfilev1.Command) error {

	eventErrors := ""

	switch {
	case len(events.PreStart) > 0:
		klog.V(2).Info("Validating preStart events")
		if preStartErr := isEventValid(events.PreStart, "preStart", commands); preStartErr != nil {
			eventErrors += fmt.Sprintf("\n%s", preStartErr.Error())
		}
		fallthrough
	case len(events.PostStart) > 0:
		klog.V(2).Info("Validating postStart events")
		if postStartErr := isEventValid(events.PostStart, "postStart", commands); postStartErr != nil {
			eventErrors += fmt.Sprintf("\n%s", postStartErr.Error())
		}
		fallthrough
	case len(events.PreStop) > 0:
		klog.V(2).Info("Validating preStop events")
		if preStopErr := isEventValid(events.PreStop, "preStop", commands); preStopErr != nil {
			eventErrors += fmt.Sprintf("\n%s", preStopErr.Error())
		}
		fallthrough
	case len(events.PostStop) > 0:
		klog.V(2).Info("Validating postStop events")
		if postStopErr := isEventValid(events.PostStop, "postStop", commands); postStopErr != nil {
			eventErrors += fmt.Sprintf("\n%s", postStopErr.Error())
		}
	}

	// if there is any validation error, return it
	if len(eventErrors) > 0 {
		return fmt.Errorf("devfile events validation error: %s", eventErrors)
	}

	return nil
}

// isEventValid checks if events belonging to a specific event type are valid ie; event should map to a valid devfile command
func isEventValid(eventNames []string, eventType string, commands map[string]devfilev1.Command) error {
	var invalidEvents []string

	for _, eventName := range eventNames {
		if _, ok := commands[strings.ToLower(eventName)]; !ok {
			klog.V(2).Infof("%s type event %s does not map to a valid devfile command", eventType, eventName)
			invalidEvents = append(invalidEvents, eventName)
		}
	}

	if len(invalidEvents) > 0 {
		klog.V(2).Infof("errors found for event type %s", eventType)
		eventErrors := fmt.Sprintf("\n%s does not map to a valid devfile command", strings.Join(invalidEvents, ", "))
		return &InvalidEventError{eventType: eventType, errorMsg: eventErrors}
	}

	return nil
}
