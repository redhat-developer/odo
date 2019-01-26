package plugin

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/metacosm/odo-event-api/odo/api/events"
	"os"
	"os/exec"
	"path"
	"strings"
)

func NewListener(pluginPath string) (events.Listener, error) {
	// get the listener name and register it in the plugin map
	listenerName, err := getListenerNameFromPath(pluginPath)
	if err != nil {
		return nil, err
	}

	// Create an hclog.Logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	// We're a host! Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         map[string]plugin.Plugin{listenerName: &ListenerGoPlugin{}},
		Cmd:             exec.Command(pluginPath),
		Logger:          logger,
		Managed:         true,
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(listenerName)
	if err != nil {
		return nil, err
	}

	listener, ok := raw.(events.Listener)
	if !ok {
		return listener, fmt.Errorf("plugin at %s is not a listener", pluginPath)
	}

	return listener, nil
}

func getListenerNameFromPath(pluginPath string) (string, error) {
	_, file := path.Split(pluginPath)
	index := strings.LastIndex(file, ".listener.odo")
	if index < 0 {
		return "", fmt.Errorf("plugin at %s doesn't follow the '<name>.listener.odo' format", pluginPath)
	}

	return file[:index], nil
}

func CleanUp() {
	plugin.CleanupClients()
}
