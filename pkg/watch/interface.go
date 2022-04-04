package watch

import (
	"context"
	"io"
)

type Client interface {
	// WatchAndPush watches the component under the context directory and triggers Push if there are any changes
	// It also listens on ctx's Done channel to trigger cleanup when indicated to do so
	WatchAndPush(out io.Writer, parameters WatchParameters, ctx context.Context) error
}
