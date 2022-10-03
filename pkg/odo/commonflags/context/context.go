package context

import (
	"context"

	"github.com/redhat-developer/odo/pkg/odo/commonflags"
)

const (
	OutputFlag = "outputFlag"
)

type (
	outputKeyType struct{}
	runOnKeyType  struct{}
)

var (
	outputKey outputKeyType
	runOnKey  runOnKeyType
)

// WithOutput sets the value for the output flag (-o) in ctx
func WithOutput(ctx context.Context, val bool) context.Context {
	return context.WithValue(ctx, outputKey, val)
}

// GetOutput gets value of output flag (-o) in ctx
func GetOutput(ctx context.Context) bool {
	value := ctx.Value(outputKey)
	if cast, ok := value.(bool); ok {
		return cast
	}
	return false
}

// WithRunOn sets the value for the run-on flag in ctx
func WithRunOn(ctx context.Context, val string) context.Context {
	return context.WithValue(ctx, runOnKey, val)
}

// GetRunOn gets value of run-on flag in ctx
func GetRunOn(ctx context.Context) string {
	value := ctx.Value(runOnKey)
	if cast, ok := value.(string); ok {
		return cast
	}
	return commonflags.RunOnDefault
}
