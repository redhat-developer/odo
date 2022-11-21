package watch

import (
	"context"
	"io"
)

type Client interface {
	// WatchAndPush watches the component under the context directory and triggers Push if there are any changes
	// It also listens on ctx's Done channel to trigger cleanup when indicated to do so
	// componentStatus is a variable to store the status of the component, and that will be exchanged between
	// parts of code (unfortunately, tthere is no place to store the status of the component in some Kubernetes resource
	// as it is generally done for a Kubernetes resource)
	WatchAndPush(out io.Writer, parameters WatchParameters, ctx context.Context, componentStatus ComponentStatus) error
}
