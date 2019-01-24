package plugin

import (
	plugin2 "github.com/metacosm/odo-event-api/odo/api/events/plugin"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/events"
)

func InitPluginAt(path string) {
	bus := events.GetEventBus()

	listener, err := plugin2.NewListener(path)
	if err != nil {
		// ignore plugin
		log.Errorf("Ignoring plugin: %v", err)
	} else {
		bus.RegisterToAll(listener)
	}
}

func CleanPlugins() {
	plugin2.CleanUp()
}
