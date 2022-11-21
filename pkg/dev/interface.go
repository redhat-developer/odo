package dev

import (
	"context"
	"io"
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
	// if RandomPorts is set, will port forward on random local ports, else uses ports starting at 40001
	RandomPorts bool
	// if WatchFiles is set, files changes will trigger a new sync to the container
	WatchFiles bool
	// Variables to override in the Devfile
	Variables map[string]string
}

type Client interface {
	// Start the resources defined in context's Devfile on the platform. It then pushes the files in path to the container.
	// It then watches for any changes to the files under path.
	// It logs messages and errors to out and errOut.
	Start(
		ctx context.Context,
		out io.Writer,
		errOut io.Writer,
		options StartOptions,
	) error

	// CleanupResources deletes the component created using the context's devfile and writes any outputs to out
	CleanupResources(ctx context.Context, out io.Writer) error
}
