package context

import (
	"context"

	"github.com/redhat-developer/odo/pkg/config"
)

type contextKey struct{}

var key = contextKey{}

// WithEnvConfig sets the environment configuration in ctx
func WithEnvConfig(ctx context.Context, val config.Configuration) context.Context {
	return context.WithValue(ctx, key, val)
}

// GetEnvConfig returns the environment configuration from ctx
func GetEnvConfig(ctx context.Context) config.Configuration {
	value := ctx.Value(key)
	if cast, ok := value.(config.Configuration); ok {
		return cast
	}
	panic("GetEnvConfig can be called only after WithEnvConfig has been called")
}
