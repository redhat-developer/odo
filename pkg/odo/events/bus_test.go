package events

import (
	"github.com/metacosm/odo-event-api/odo/api/events"
	"reflect"
	"testing"
)

type testListener struct{}

func (testListener) OnEvent(event events.Event) error {
	panic("implement me")
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
	bus.RegisterSingle(name, eventType, listener)
	if !reflect.DeepEqual(bus.getListenersFor(name, eventType)[0], listener) {
		t.Error("failed, expected bus listeners to contain registered listener")
	}
}
