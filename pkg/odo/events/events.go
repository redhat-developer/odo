package events

type EventBus struct {
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

var bus *EventBus = &EventBus{}

func GetEventBus() *EventBus {
	return bus
}

func (bus *EventBus) DispatchEvent(event Event) (err error) {
	return
}
