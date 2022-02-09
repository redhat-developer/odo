package watch

import (
	"github.com/redhat-developer/odo/pkg/kclient"
	"io"
)

type Client interface {
	// TODO: Add docs
	DevfileWatchAndPush(out io.Writer, parameters WatchParameters) error
	// TODO: Add docs
	WatchAndPush(client kclient.ClientInterface, out io.Writer, parameters WatchParameters) error
}
