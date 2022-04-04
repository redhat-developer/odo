package watch

import (
	"context"
	"io"
)

type Client interface {
	// WatchAndPush watches the component under the context directory and triggers Push if there are any changes
	WatchAndPush(out io.Writer, parameters WatchParameters, ctx context.Context) error
}
