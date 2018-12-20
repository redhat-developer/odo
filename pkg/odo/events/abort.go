package events

import (
	"fmt"
	"reflect"
)

type EventCausedAbortError struct {
	Listener Listener
	Source   Event
	cause    error
}

func (e *EventCausedAbortError) Error() string {
	return fmt.Sprintf("listener %s aborted processing on event %v, cause: %v", e.Listener.Name(), e.Source, e.cause)
}

func (e *EventCausedAbortError) Cause() error {
	return e.cause
}

func NewEventCausedAbortError(listener Listener, event Event, cause error) *EventCausedAbortError {
	return &EventCausedAbortError{
		Listener: listener,
		Source:   event,
		cause:    cause,
	}
}

func IsEventCausedAbort(err error) bool {
	return reflect.TypeOf(err) == reflect.TypeOf(EventCausedAbortError{})
}
