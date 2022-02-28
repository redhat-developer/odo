package watch

import (
	"io"
)

type Client interface {
	// WatchAndPush watches the component under the context directory and triggers Push if there are any changes
	WatchAndPush(out io.Writer, parameters WatchParameters) error
}
