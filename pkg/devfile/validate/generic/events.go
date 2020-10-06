package generic

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"k8s.io/klog"
)

// IsEventValid checks if events belonging to a specific event type are valid ie; event should map to a valid devfile command
func IsEventValid(eventNames []string, eventType string, commands map[string]common.DevfileCommand, components []common.DevfileComponent) error {
	eventErrorMsg := make(map[string][]string)
	eventErrors := ""

	for _, eventName := range eventNames {
		isEventPresent := false

		if _, ok := commands[strings.ToLower(eventName)]; ok {
			isEventPresent = true
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
