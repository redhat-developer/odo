package watch

import (
	"io"
)

type Client interface {
	//
	DevfileWatchAndPush(out io.Writer, parameters WatchParameters) error
	//
	WatchAndPush(out io.Writer, parameters WatchParameters) error
}
