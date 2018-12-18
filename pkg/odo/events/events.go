package events

import (
	"fmt"
	"github.com/spf13/cobra"
)

type typesToListeners map[EventType][]Listener

type EventBus struct {
	listeners    map[string]typesToListeners
	allListeners []Listener
}

type Listener interface {
	OnEvent(event Event) error
}

type EventType int

func (t EventType) String() string {
	return eventTypes[t]
}

var eventTypes = []string{
	"Unknown",
	"PreRun",
	"PostRun",
}

const (
	Unknown EventType = iota
	PreRun
	PostRun
)

type Event struct {
	Name string
	Type EventType
}

func (e Event) String() string {
	return e.Name + "->" + e.Type.String()
}

var bus = &EventBus{
	allListeners: make([]Listener, 0, 5),
}

func GetEventBus() *EventBus {
	return bus
}

func (bus *EventBus) RegisterToAll(listener Listener) {
	bus.allListeners = append(bus.allListeners, listener)
}

func (bus *EventBus) Register(event string, eventType EventType, listener Listener) {
	listenersForEvent, ok := bus.listeners[event]
	if !ok {
		listenersForEvent = make(typesToListeners, 10)
		bus.listeners[event] = listenersForEvent
	}

	listenersForType, ok := listenersForEvent[eventType]
	if !ok {
		listenersForType = make([]Listener, 0, 10)
	}

	listenersForEvent[eventType] = append(listenersForType, listener)
}

func (bus *EventBus) DispatchEvent(event Event) (err error) {
	errors := make([]error, 0, 10)
	listenersForEvent, ok := bus.listeners[event.Name]
	if ok {
		listenersForType, ok := listenersForEvent[event.Type]
		if ok {
			for i := range listenersForType {
				listener := listenersForType[i]
				err := listener.OnEvent(event)
				if err != nil {
					errors = append(errors, err)
				}
			}
		}
	}

	for i := range bus.allListeners {
		err := bus.allListeners[i].OnEvent(event)
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		msg := ""
		for e := range errors {
			msg = msg + fmt.Sprintf("\n%v", errors[e])
		}
		return fmt.Errorf("%d error(s) occurred while processing event %v: %s", len(errors), event, msg)
	}
	return
}

func EventNameFrom(cmd *cobra.Command) string {
	if cmd.HasParent() {
		return EventNameFrom(cmd.Parent()) + ":" + cmd.Name()
	}
	return cmd.Name()
}

type tracer struct{}

func (t tracer) OnEvent(event Event) error {
	fmt.Printf("got event %v\n", event)
	return nil
}

/*func init() {
	GetEventBus().RegisterToAll(tracer{})
}*/
