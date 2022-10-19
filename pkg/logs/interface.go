package logs

import "context"

type Client interface {
	// GetLogsForMode gets logs of the containers for the specified mode (Dev, Deploy or both) of the provided
	// component name and namespace. It returns Events which has multiple channels. Logs are put on the
	// Events.Logs channel and errors on Events.Err. Events.Done channel is populated to indicate that all Pods' logs
	// have been fetched.
	// The accepted values for mode are ComponentDevMode, ComponentDeployMode and ComponentAnyMode
	// found in the pkg/labels package.
	// Setting follow boolean to true helps follow/tail the logs of the pods.
	GetLogsForMode(
		ctx context.Context,
		mode string,
		componentName string,
		namespace string,
		follow bool,
	) (Events, error)
}
