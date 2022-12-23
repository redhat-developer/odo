package context

import (
	"context"
)

const (
	OutputFlag = "outputFlag"
)

type (
	outputKeyType    struct{}
	platformKeyType  struct{}
	variablesKeyType struct{}
)

var (
	outputKey    outputKeyType
	platformKey  platformKeyType
	variablesKey variablesKeyType
)

// WithJsonOutput sets the value for the output flag (-o) in ctx
func WithJsonOutput(ctx context.Context, val bool) context.Context {
	return context.WithValue(ctx, outputKey, val)
}

// IsJsonOutput gets value of output flag (-o) in ctx
func IsJsonOutput(ctx context.Context) bool {
	value := ctx.Value(outputKey)
	if cast, ok := value.(bool); ok {
		return cast
	}
	return false
}

// WithPlatform sets the value for the platform flag in ctx
func WithPlatform(ctx context.Context, val string) context.Context {
	return context.WithValue(ctx, platformKey, val)
}

// GetPlatform gets value of platform flag in ctx
func GetPlatform(ctx context.Context, defaultValue string) string {
	value := ctx.Value(platformKey)
	if cast, ok := value.(string); ok {
		return cast
	}
	return defaultValue
}

// WithVariables sets the value for the --var-file and --var flags in ctx
func WithVariables(ctx context.Context, val map[string]string) context.Context {
	return context.WithValue(ctx, variablesKey, val)
}

// GetVariables gets values of --var-file and --var flags in ctx
func GetVariables(ctx context.Context) map[string]string {
	value := ctx.Value(variablesKey)
	if cast, ok := value.(map[string]string); ok {
		return cast
	}
	return map[string]string{}
}
