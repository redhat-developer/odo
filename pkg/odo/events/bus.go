package events

import (
	"fmt"
	api "github.com/metacosm/odo-event-api/odo/api/events"
	"github.com/spf13/cobra"
)

type typesToListeners map[api.EventType][]api.Listener

type EventBus struct {
	listeners    map[string]typesToListeners
	allListeners []api.Listener
}

var bus = &EventBus{
	allListeners: make([]api.Listener, 0, 5),
}

func GetEventBus() *EventBus {
	return bus
}

func EventNameFrom(cmd *cobra.Command) string {
	if cmd.HasParent() {
		return EventNameFrom(cmd.Parent()) + ":" + cmd.Name()
	}
	return cmd.Name()
}

func DispatchEvent(cmd *cobra.Command, eventType api.EventType, payload interface{}) error {
	eventBus := GetEventBus()
	err := eventBus.DispatchEvent(api.Event{
		Name:    EventNameFrom(cmd),
		Type:    eventType,
		Payload: payload,
	})

	return err
}

func (bus *EventBus) RegisterToAll(listener api.Listener) {
	bus.allListeners = append(bus.allListeners, listener)
}

type Subscription struct {
	Listener        api.Listener
	SupportedEvents map[string]api.EventType
}

func (bus *EventBus) Register(subscription Subscription) {
	for k, v := range subscription.SupportedEvents {
		bus.RegisterSingle(k, v, subscription.Listener)
	}
}

func (bus *EventBus) RegisterSingle(event string, eventType api.EventType, listener api.Listener) {
	listenersForEvent, ok := bus.listeners[event]
	if !ok {
		listenersForEvent = make(typesToListeners, 10)
		bus.listeners[event] = listenersForEvent
	}

	listenersForType, ok := listenersForEvent[eventType]
	if !ok {
		listenersForType = make([]api.Listener, 0, 10)
	}

	listenersForEvent[eventType] = append(listenersForType, listener)
}

func (bus *EventBus) getListenersFor(name string, eventType api.EventType) []api.Listener {
	listenersForEvent, ok := bus.listeners[name]
	if ok {
		listenersForType, ok := listenersForEvent[eventType]
		if ok {
			return listenersForType
		}
	}
	return []api.Listener{}
}

func (bus *EventBus) DispatchEvent(event api.Event) (err error) {
	errors := make([]error, 0, 10)
	var abort bool
	processedListeners := make([]api.Listener, 0, 10)

	listeners := bus.getListenersFor(event.Name, event.Type)
	for i := range listeners {
		listener := listeners[i]
		err := listener.OnEvent(event)
		if err != nil {
			if api.IsEventCausedAbort(err) {
				abort = true
				return err
			}
			errors = append(errors, err)
		}

		processedListeners = append(processedListeners, listener)
	}

	for i := range bus.allListeners {
		listener := bus.allListeners[i]
		err := listener.OnEvent(event)
		if err != nil {
			if api.IsEventCausedAbort(err) {
				abort = true
				return err
			}
			errors = append(errors, err)
		}

		processedListeners = append(processedListeners, listener)
	}

	defer revertProcessedListenersOnAbort(abort, err, processedListeners)

	if len(errors) > 0 {
		msg := ""
		for e := range errors {
			msg = msg + fmt.Sprintf("\n%v", errors[e])
		}
		return fmt.Errorf("%d error(s) occurred while processing event %v: %s", len(errors), event, msg)
	}
	return
}

func revertProcessedListenersOnAbort(abort bool, err error, listeners []api.Listener) {
	if abort {

		for i := range listeners {
			listeners[i].OnAbort(err.(api.EventCausedAbortError))
		}
	}
}
