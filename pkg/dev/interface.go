package dev

import (
	"context"
	"io"

	"github.com/devfile/library/pkg/devfile/parser"

	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/watch"
)

type Client interface {
	// Start the resources in devfileObj on the namespace. It then pushes the files in path to the container.
	// If debug is true, executes the debug command, or the run command by default.
	// If buildCommand is set, this will look up the specified build command in the Devfile. Otherwise, it uses the default one.
	// If runCommand is set, this will look up the specified run command in the Devfile and execute it. Otherwise, it uses the default one.
	// Returns the status of the started component
	Start(
		devfileObj parser.DevfileObj,
		componentName string,
		namespace string,
		ignorePaths []string,
		path string,
		debug bool,
		buildCommand string,
		runCommand string,
		randomPorts bool,
		errOut io.Writer,
		fs filesystem.Filesystem,
	) (watch.ComponentStatus, error)

	// Watch watches for any changes to the files under path while ignoring the files/directories in ignorePaths.
	// It logs messages to out and uses the Handler h to perform push operation when anything changes in path.
	// It uses devfileObj to notify user to restart odo dev if they change endpoint information in the devfile.
	// If debug is true, the debug command will be started after a sync, or the run command by default.
	// If buildCommand is set, this will look up the specified build command in the Devfile. Otherwise, it uses the default one.
	// If runCommand is set, this will look up the specified run command in the Devfile and execute it. Otherwise, it uses the default one.
	// componentStatus is the status returned from the call to the Start Method
	Watch(
		devfilePath string,
		devfileObj parser.DevfileObj,
		componentName string,
		path string,
		ignorePaths []string,
		out io.Writer,
		h Handler,
		ctx context.Context,
		debug bool,
		buildCommand string,
		runCommand string,
		variables map[string]string,
		randomPorts bool,
		watchFiles bool,
		errOut io.Writer,
		componentStatus watch.ComponentStatus,
	) error
}

type Handler interface {
	RegenerateAdapterAndPush(adapters.PushParameters, watch.WatchParameters, *watch.ComponentStatus) error
}
