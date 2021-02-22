package validation

import (
	"fmt"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

const (
	preStart  = "preStart"
	postStart = "postStart"
	preStop   = "preStop"
	postStop  = "postStop"
)

// ValidateEvents validates all the devfile events
func ValidateEvents(events v1alpha2.Events, commands []v1alpha2.Command) error {

	eventErrors := ""

	commandMap := getCommandsMap(commands)

	switch {
	case len(events.PreStart) > 0:
		if preStartErr := isEventValid(events.PreStart, preStart, commandMap); preStartErr != nil {
			eventErrors += fmt.Sprintf("\n%s", preStartErr.Error())
		}
		fallthrough
	case len(events.PostStart) > 0:
		if postStartErr := isEventValid(events.PostStart, postStart, commandMap); postStartErr != nil {
			eventErrors += fmt.Sprintf("\n%s", postStartErr.Error())
		}
		fallthrough
	case len(events.PreStop) > 0:
		if preStopErr := isEventValid(events.PreStop, preStop, commandMap); preStopErr != nil {
			eventErrors += fmt.Sprintf("\n%s", preStopErr.Error())
		}
		fallthrough
	case len(events.PostStop) > 0:
		if postStopErr := isEventValid(events.PostStop, postStop, commandMap); postStopErr != nil {
			eventErrors += fmt.Sprintf("\n%s", postStopErr.Error())
		}
	}

	// if there is any validation error, return it
	if len(eventErrors) > 0 {
		return fmt.Errorf("devfile events validation error: %s", eventErrors)
	}

	return nil
}

// isEventValid checks if events belonging to a specific event type are valid ie;
// 1. event should map to a valid devfile command
// 2. preStart and postStop events should either map to an apply command or a composite command with apply commands
// 3. postStart and preStop events should either map to an exec command or a composite command with exec commands
func isEventValid(eventNames []string, eventType string, commandMap map[string]v1alpha2.Command) error {
	var invalidCommand, invalidApplyEvents, invalidExecEvents []string

	for _, eventName := range eventNames {
		command, ok := commandMap[strings.ToLower(eventName)]
		if !ok { // check if event is in the list of devfile commands
			invalidCommand = append(invalidCommand, eventName)
			continue
		}

		switch eventType {
		case preStart, postStop:
			// check if the event is either an apply command or a composite of apply commands
			if command.Apply == nil && command.Composite == nil {
				invalidApplyEvents = append(invalidApplyEvents, eventName)
			} else if command.Composite != nil {
				invalidApplyEvents = append(invalidApplyEvents, validateCompositeEvents(*command.Composite, eventName, eventType, commandMap)...)
			}
		case postStart, preStop:
			// check if the event is either an exec command or a composite of exec commands
			if command.Exec == nil && command.Composite == nil {
				invalidExecEvents = append(invalidExecEvents, eventName)
			} else if command.Composite != nil {
				invalidExecEvents = append(invalidExecEvents, validateCompositeEvents(*command.Composite, eventName, eventType, commandMap)...)
			}
		}
	}

	var eventErrors string
	var err error

	if len(invalidCommand) > 0 {
		eventErrors = fmt.Sprintf("\n%s does not map to a valid devfile command", strings.Join(invalidCommand, ", "))
	}

	if len(invalidApplyEvents) > 0 {
		eventErrors += fmt.Sprintf("\n%s should either map to an apply command or a composite command with apply commands", strings.Join(invalidApplyEvents, ", "))
	}

	if len(invalidExecEvents) > 0 {
		eventErrors += fmt.Sprintf("\n%s should either map to an exec command or a composite command with exec commands", strings.Join(invalidExecEvents, ", "))
	}

	if len(eventErrors) != 0 {
		err = &InvalidEventError{eventType: eventType, errorMsg: eventErrors}
	}

	return err
}

// validateCompositeEvents checks if a composite subcommands are
// 1. apply commands for preStart and postStop
// 2. exec commands for postStart and preStop
func validateCompositeEvents(composite v1alpha2.CompositeCommand, eventName, eventType string, commandMap map[string]v1alpha2.Command) []string {
	var invalidEvents []string

	switch eventType {
	case preStart, postStop:
		for _, subCommand := range composite.Commands {
			if command, ok := commandMap[subCommand]; ok && command.Apply == nil {
				invalidEvents = append(invalidEvents, eventName)
			}
		}
	case postStart, preStop:
		for _, subCommand := range composite.Commands {
			if command, ok := commandMap[subCommand]; ok && command.Exec == nil {
				invalidEvents = append(invalidEvents, eventName)
			}
		}
	}

	return invalidEvents
}
