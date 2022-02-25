package watch

import (
	"github.com/redhat-developer/odo/pkg/kclient"
	"io"
)

type Client interface {
	// WatchAndPush watches the component under the context directory and triggers Push if there are any changes
	WatchAndPush(client kclient.ClientInterface, out io.Writer, parameters WatchParameters) error
}
