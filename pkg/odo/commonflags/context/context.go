package context

import "context"

const (
	OutputFlag = "outputFlag"
)

type outputKeyType struct{}

var outputKey outputKeyType

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
