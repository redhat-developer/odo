package plugin

import (
	"fmt"
	api "github.com/metacosm/odo-event-api/odo/api/events"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/events"
	"os"
	"plugin"
)

func InitPluginAt(path string) {
	bus := events.GetEventBus()

	// Open the plugin library
	plug, err := plugin.Open(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Look up the exporter Listener variable
	candidate, err := plug.Lookup("Listener")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Assert that Listener is indeed a Listener :)
	listener, ok := candidate.(api.Listener)
	if !ok {
		log.Error("exported Listener variable is not implementing the Listener interface")
		os.Exit(1)
	}

	bus.RegisterToAll(listener)
}
