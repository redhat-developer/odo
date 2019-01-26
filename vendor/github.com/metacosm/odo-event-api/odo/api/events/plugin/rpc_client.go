package plugin

import (
	api "github.com/metacosm/odo-event-api/odo/api/events"
	"net/rpc"
)

type ListenerRPC struct{ client *rpc.Client }

func (g *ListenerRPC) OnEvent(event api.Event) error {
	var foo string
	err := g.client.Call("Plugin.OnEvent", event, &foo)
	if err != nil {
		return err
	}

	return nil
}

func (g *ListenerRPC) OnAbort(abortError api.EventCausedAbortError) {
	var foo string
	err := g.client.Call("Plugin.OnAbort", abortError, &foo)
	if err != nil {
		panic(err)
	}
}

func (g *ListenerRPC) Name() string {
	var name string
	err := g.client.Call("Plugin.Name", new(interface{}), &name)
	if err != nil {
		panic(err)
	}
	return name
}
