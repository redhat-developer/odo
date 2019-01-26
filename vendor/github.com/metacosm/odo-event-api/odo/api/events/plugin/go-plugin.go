package plugin

import (
	"github.com/hashicorp/go-plugin"
	api "github.com/metacosm/odo-event-api/odo/api/events"
	"net/rpc"
)

// This is the implementation of plugin.Plugin so we can serve/consume this
//
// This has two methods: Server must return an RPC server for this plugin
// type. We construct a ListenerRPCServer for this.
//
// Client must return an implementation of our interface that communicates
// over an RPC client. We return GreeterRPC for this.
//
// Ignore MuxBroker. That is used to create more multiplexed streams on our
// plugin connection and is a more advanced use case.
type ListenerGoPlugin struct {
	// Impl Injection
	Impl api.Listener
}

func (p *ListenerGoPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &ListenerRPCServer{Impl: p.Impl}, nil
}

func (ListenerGoPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ListenerRPC{client: c}, nil
}
