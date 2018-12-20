package events

import "fmt"

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

/*func init() {
	GetEventBus().RegisterToAll(tracer{})
}*/
