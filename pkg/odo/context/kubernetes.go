package context

import (
	"context"
)

type (
	namespaceKeyType struct{}
)

var (
	namespaceKey namespaceKeyType
)

// WithNamespace sets the value of the current namespace in ctx
func WithNamespace(ctx context.Context, val string) context.Context {
	return context.WithValue(ctx, namespaceKey, val)
}

// GetNamespace gets the namespace value in ctx
func GetNamespace(ctx context.Context) string {
	value := ctx.Value(namespaceKey)
	if cast, ok := value.(string); ok {
		return cast
	}
	return ""
}
