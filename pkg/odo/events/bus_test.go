package events

import (
	"github.com/metacosm/odo-event-api/odo/api/events"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

var received events.Event

type testListener struct {
}

func (t testListener) OnEvent(event events.Event) error {
	received = event
	return nil
}

func (testListener) OnAbort(abortError events.EventCausedAbortError) {
	panic("implement me")
}

func (testListener) Name() string {
	panic("implement me")
}

func TestEventBus_RegisterSingle(t *testing.T) {
	listener := testListener{}
	name := "foo"
	eventType := events.PostComplete

	// initial state should be empty
	assert.Equal(t, len(bus.listeners), 0)
	assert.NotContains(t, bus.listeners, name)

	// we should now only one listener
	bus.RegisterSingle(name, eventType, listener)
	assert.Equal(t, len(bus.listeners), 1)
	assert.Contains(t, bus.listeners, name)
	assert.Equal(t, len(bus.listeners[name]), 1)
	assert.Contains(t, bus.listeners[name], eventType)
	if !reflect.DeepEqual(bus.getListenersFor(name, eventType)[0], listener) {
		t.Error("failed, expected bus listeners to contain registered listener")
	}

	// we should still have one listener
	bus.RegisterSingle(name, eventType, nil)
	assert.Equal(t, len(bus.listeners), 1)
	assert.Contains(t, bus.listeners, name)
	assert.Equal(t, len(bus.listeners[name]), 1)
	assert.Contains(t, bus.listeners[name], eventType)

	// register a listener for a different event and type and check that it's properly registered
	bar := "bar"
	eventType2 := events.PostValidate
	bus.RegisterSingle(bar, eventType2, listener)
	assert.Equal(t, len(bus.listeners), 2)
	assert.Contains(t, bus.listeners, name)
	assert.Contains(t, bus.listeners, bar)
	assert.Equal(t, len(bus.listeners[name]), 1)
	assert.Equal(t, len(bus.listeners[bar]), 1)
	assert.Contains(t, bus.listeners[name], eventType)
	assert.Contains(t, bus.listeners[bar], eventType2)
	if !reflect.DeepEqual(bus.getListenersFor(bar, eventType2)[0], listener) {
		t.Error("failed, expected bus listeners to contain registered listener")
	}

	// reset state
	delete(bus.listeners, name)
	delete(bus.listeners, bar)
	assert.Equal(t, len(bus.listeners), 0)
}

func TestDispatchEvent(t *testing.T) {
	foo := "foo"
	eventType := events.PostValidate
	event := events.Event{
		Name: foo,
		Type: eventType,
	}

	// received shouldn't changed when no listeners are registered
	err := bus.DispatchEvent(event)
	if err != nil {
		t.Errorf("failed, got unexpected error: %v", err)
	}
	assert.EqualValues(t, received, events.Event{})

	// check that received is indeed equal to the dispatched event
	listener := testListener{}
	bus.RegisterSingle(foo, eventType, listener)
	err = bus.DispatchEvent(event)
	if err != nil {
		t.Errorf("failed, got unexpected error: %v", err)
	}
	assert.EqualValues(t, event, received)

	// reset state
	delete(bus.listeners, foo)
	received = events.Event{}
}
