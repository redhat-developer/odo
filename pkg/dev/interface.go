package dev

import (
	"context"
	"io"

	"github.com/redhat-developer/odo/pkg/api"
)

type StartOptions struct {
	// IgnorePaths are files/directories to ignore when pushing files to the container.
	IgnorePaths []string
	// If Debug is true, executes the debug command, or the run command by default.
	Debug bool
	// If BuildCommand is set, this will look up the specified build command in the Devfile. Otherwise, it uses the default one.
	BuildCommand string
	// If RunCommand is set, this will look up the specified run command in the Devfile and execute it. Otherwise, it uses the default one.
	RunCommand string
	// If DebugCommand is set, this will look up the specified debug command in the Devfile and execute it. Otherwise, it uses the default one.
	DebugCommand string
	// if RandomPorts is set, will port forward on random local ports, else uses ports starting at 20001
	RandomPorts bool
	// CustomForwardedPorts define custom ports for port forwarding
	CustomForwardedPorts []api.ForwardedPort
	// CustomAddress defines a custom local address for port forwarding; default value is 127.0.0.1
	CustomAddress string
	// if WatchFiles is set, files changes will trigger a new sync to the container
	WatchFiles bool
	// IgnoreLocalhost indicates whether to proceed with port-forwarding regardless of any container ports being bound to the container loopback interface.
	// Applicable to Podman only.
	IgnoreLocalhost bool
	// ForwardLocalhost is a flag indicating if we inject a side container that will make port-forwarding work with container apps listening on the loopback interface.
	// Applicable to Podman only.
	ForwardLocalhost bool
	// Variables to override in the Devfile
	Variables map[string]string

	Out    io.Writer
	ErrOut io.Writer
}

type Client interface {
	// Start the resources defined in context's Devfile on the platform. It then pushes the files in path to the container.
	// It then watches for any changes to the files under path.
	// It logs messages and errors to out and errOut.
	Start(
		ctx context.Context,
		options StartOptions,
	) error

	Run(
		ctx context.Context,
		commandName string,
	) error

	// CleanupResources deletes the component created using the context's devfile and writes any outputs to out
	CleanupResources(ctx context.Context, out io.Writer) error
}
