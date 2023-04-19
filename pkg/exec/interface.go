package exec

import (
	"context"
	"io"
)

type Client interface {
	// ExecuteCommand executes the given command in the pod's container,
	// writing the output to the specified respective pipe writers
	ExecuteCommand(ctx context.Context, command []string, podName string, containerName string, show bool, stdoutWriter *io.PipeWriter, stderrWriter *io.PipeWriter) (stdout []string, stderr []string, err error)
}
