package validation

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
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
func ValidateEvents(events v1alpha2.Events, commands []v1alpha2.Command) (err error) {

	commandMap := getCommandsMap(commands)

	switch {
	case len(events.PreStart) > 0:
		if preStartErr := isEventValid(events.PreStart, preStart, commandMap); preStartErr != nil {
			err = multierror.Append(err, preStartErr)
		}
		fallthrough
	case len(events.PostStart) > 0:
		if postStartErr := isEventValid(events.PostStart, postStart, commandMap); postStartErr != nil {
			err = multierror.Append(err, postStartErr)
		}
		fallthrough
	case len(events.PreStop) > 0:
		if preStopErr := isEventValid(events.PreStop, preStop, commandMap); preStopErr != nil {
			err = multierror.Append(err, preStopErr)
		}
		fallthrough
	case len(events.PostStop) > 0:
		if postStopErr := isEventValid(events.PostStop, postStop, commandMap); postStopErr != nil {
			err = multierror.Append(err, postStopErr)
		}
	}

	return err
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

	var err error
	var eventErrorsList []string

	if len(invalidCommand) > 0 {
		eventErrorsList = append(eventErrorsList, fmt.Sprintf("%s does not map to a valid devfile command", strings.Join(invalidCommand, ", ")))
	}

	if len(invalidApplyEvents) > 0 {
		eventErrorsList = append(eventErrorsList, fmt.Sprintf("%s should either map to an apply command or a composite command with apply commands", strings.Join(invalidApplyEvents, ", ")))
	}

	if len(invalidExecEvents) > 0 {
		eventErrorsList = append(eventErrorsList, fmt.Sprintf("%s should either map to an exec command or a composite command with exec commands", strings.Join(invalidExecEvents, ", ")))
	}

	if len(eventErrorsList) != 0 {
		eventErrors := fmt.Sprintf("\n%s", strings.Join(eventErrorsList, "\n"))
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
