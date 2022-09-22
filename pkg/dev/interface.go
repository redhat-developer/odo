package dev

import (
	"context"
	"io"

	"github.com/devfile/library/pkg/devfile/parser"
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
	// if RandomPorts is set, will port forward on random local ports, else uses ports starting at 40001
	RandomPorts bool
	// if WatchFiles is set, files changes will trigger a new sync to the container
	WatchFiles bool
	// Variables to override in the Devfile
	Variables map[string]string
}

type Client interface {
	// Start the resources in devfileObj on the namespace. It then pushes the files in path to the container.
	// It then watches for any changes to the files under path.
	// It logs messages and errors to out and errOut.
	Start(
		ctx context.Context,
		devfileObj parser.DevfileObj,
		componentName string,
		path string,
		devfilePath string,
		out io.Writer,
		errOut io.Writer,
		options StartOptions,
	) error
}
