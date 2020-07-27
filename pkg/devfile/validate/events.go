package validate

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"k8s.io/klog"
)

// validateEvents validates all the devfile events
func validateEvents(events common.DevfileEvents, commands []common.DevfileCommand) error {

	eventErrors := ""

	switch {
	case len(events.PreStart) > 0:
		klog.V(4).Info("Validating preStart events")
		if preStartErr := isEventValid(events.PreStart, "preStart", commands); preStartErr != nil {
			eventErrors += fmt.Sprintf("\n%s", preStartErr.Error())
		}
		fallthrough
	case len(events.PostStart) > 0:
		klog.V(4).Info("Validating postStart events")
		if postStartErr := isEventValid(events.PostStart, "postStart", commands); postStartErr != nil {
			eventErrors += fmt.Sprintf("\n%s", postStartErr.Error())
		}
		fallthrough
	case len(events.PreStop) > 0:
		klog.V(4).Info("Validating preStop events")
		if preStopErr := isEventValid(events.PreStop, "preStop", commands); preStopErr != nil {
			eventErrors += fmt.Sprintf("\n%s", preStopErr.Error())
		}
		fallthrough
	case len(events.PostStop) > 0:
		klog.V(4).Info("Validating postStop events")
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

func isEventValid(eventNames []string, eventType string, commands []common.DevfileCommand) error {
	var invalidEvents []string

	for _, eventName := range eventNames {
		isValid := false
		for _, command := range commands {
			if command.GetID() == strings.ToLower(eventName) {
				isValid = true
				break
			}
		}

		if !isValid {
			klog.V(4).Infof("%s type event %s invalid", eventType, eventName)
			invalidEvents = append(invalidEvents, eventName)
		}
	}

	if len(invalidEvents) > 0 {
		return &InvalidEventError{eventType: eventType, event: strings.Join(invalidEvents, ",")}
	}

	return nil
}
