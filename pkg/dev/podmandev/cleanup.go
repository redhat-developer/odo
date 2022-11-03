package podmandev

import (
	"context"
	"io"

	"github.com/devfile/library/pkg/devfile/parser"
)

func (o *DevClient) CleanupResources(ctx context.Context, devfileObj parser.DevfileObj, componentName string, out io.Writer) error {
	return nil
}
