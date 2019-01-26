package plugin

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	api "github.com/metacosm/odo-event-api/odo/api/events"
	"os"
)

type BasePlugin struct {
	logger   hclog.Logger
	listener api.Listener
}

func (p *BasePlugin) OnEvent(event api.Event) error {
	p.logger.Info("event: %v received by listener: %v", event, p.listener)
	return p.listener.OnEvent(event)
}
func (p *BasePlugin) OnAbort(abortError api.EventCausedAbortError) {
	p.listener.OnAbort(abortError)
}
func (p *BasePlugin) Name() string {
	return p.listener.Name()
}

func (p *BasePlugin) Serve() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         map[string]plugin.Plugin{p.listener.Name(): &ListenerGoPlugin{Impl: p.listener}},
	})
}

// handshakeConfigs are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "ODO_LISTENER_PLUGIN",
	MagicCookieValue: "odo.listener.plugin",
}

func NewPlugin(listener api.Listener) BasePlugin {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Trace,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	return BasePlugin{
		logger:   logger,
		listener: listener,
	}

}
