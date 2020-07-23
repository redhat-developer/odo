package validate

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"k8s.io/klog"
)

// Errors
var (
	ErrorInvalidEvent = "%s type event %s invalid"
)

// validateEvents validates all the devfile events
func validateEvents(events common.DevfileEvents, commands []common.DevfileCommand) error {

	var preStartErr, postStartErr, preStopErr, postStopErr error

	switch {
	case len(events.PreStart) > 0:
		klog.V(4).Info("Validating preStart events")
		preStartErr = isEventValid(events.PreStart, "preStart", commands)
		fallthrough
	case len(events.PostStart) > 0:
		klog.V(4).Info("Validating postStart events")
		postStartErr = isEventValid(events.PostStart, "postStart", commands)
		fallthrough
	case len(events.PreStop) > 0:
		klog.V(4).Info("Validating preStop events")
		preStopErr = isEventValid(events.PreStop, "preStop", commands)
		fallthrough
	case len(events.PostStop) > 0:
		klog.V(4).Info("Validating postStop events")
		postStopErr = isEventValid(events.PostStop, "postStop", commands)
	}

	eventErrors := ""
	if preStartErr != nil {
		eventErrors += fmt.Sprintf("\n%s", preStartErr.Error())
	}
	if postStartErr != nil {
		eventErrors += fmt.Sprintf("\n%s", postStartErr.Error())
	}
	if preStopErr != nil {
		eventErrors += fmt.Sprintf("\n%s", preStopErr.Error())
	}
	if postStopErr != nil {
		eventErrors += fmt.Sprintf("\n%s", postStopErr.Error())
	}

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
			if command.Exec != nil && command.Exec.Id == strings.ToLower(eventName) {
				isValid = true
				break
			}
		}

		if !isValid {
			klog.V(4).Infof(ErrorInvalidEvent, eventType, eventName)
			invalidEvents = append(invalidEvents, eventName)
		}
	}

	if len(invalidEvents) > 0 {
		return fmt.Errorf(ErrorInvalidEvent, eventType, strings.Join(invalidEvents, ","))
	}

	return nil
}
