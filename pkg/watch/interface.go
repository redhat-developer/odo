package watch

import (
	"context"
	"io"

	"github.com/devfile/library/pkg/devfile/parser"
)

type Client interface {
	// WatchAndPush watches the component under the context directory and triggers Push if there are any changes
	// It also listens on ctx's Done channel to trigger cleanup when indicated to do so
	WatchAndPush(out io.Writer, parameters WatchParameters, ctx context.Context) error
	// Cleanup deletes the component created using the devfileObj and writes any outputs to out
	Cleanup(devfileObj parser.DevfileObj, out io.Writer) error
}
