package events

import (
	"fmt"
	"os"
)

type tracer struct{}

func (t tracer) OnEvent(event Event) error {
	fmt.Printf("got event %v\n", event)
	return nil
}

func (t tracer) OnAbort(abortError *EventCausedAbortError) {
	fmt.Printf("abort: %v\n", abortError)
}

func (t tracer) Name() string {
	return "tracer"
}

func init() {
	if os.Getenv("ODO_EVENT_TRACER") == "true" {
		GetEventBus().RegisterToAll(tracer{})
	}
}
