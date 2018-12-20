package events

type EventType int

func (t EventType) String() string {
	return eventTypes[t]
}

var eventTypes = []string{
	"Unknown",
	"PreRun",
	"PostComplete",
	"PostValidate",
	"PostRun",
}

const (
	Unknown EventType = iota
	PreRun
	PostComplete
	PostValidate
	PostRun
)

type Event struct {
	Name    string
	Type    EventType
	Payload interface{}
}

func (e Event) String() string {
	return e.Name + "->" + e.Type.String()
}
