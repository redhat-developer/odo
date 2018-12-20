package events

type Listener interface {
	OnEvent(event Event) error
	OnAbort(abortError *EventCausedAbortError)
	Name() string
}
