package dev

import (
	"context"
	"io"

	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"

	"github.com/devfile/library/pkg/devfile/parser"

	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
	"github.com/redhat-developer/odo/pkg/watch"
)

type Client interface {
	// Start the resources in devfileObj on the platformContext. It then pushes the files in path to the container.
	// If debug is true, executes the debug command, or the run command by default.
	// If runCommand is set, this will look up the specified run command in the Devfile and execute it. Otherwise, it uses the default one.
	Start(
		devfileObj parser.DevfileObj,
		platformContext kubernetes.KubernetesContext,
		ignorePaths []string,
		path string,
		debug bool,
		runCommand string,
	) error

	// Watch watches for any changes to the files under path while ignoring the files/directories in ignorePaths.
	// It logs messages to out and uses the Handler h to perform push operation when anything changes in path.
	// It uses devfileObj to notify user to restart odo dev if they change endpoint information in the devfile.
	// If debug is true, the debug command will be started after a sync, or the run command by default.
	// If runCommand is set, this will look up the specified run command in the Devfile and execute it. Otherwise, it uses the default one.
	Watch(
		devfileObj parser.DevfileObj,
		path string,
		ignorePaths []string,
		out io.Writer,
		h Handler,
		ctx context.Context,
		debug bool,
		runCommand string,
		variables map[string]string,
	) error
}

type Handler interface {
	RegenerateAdapterAndPush(common.PushParameters, watch.WatchParameters) error
}
