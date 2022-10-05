package context

import "context"

type (
	applicationKeyType struct{}
)

var (
	applicationKey applicationKeyType
)

// WithApplication sets the value of the application in ctx
func WithApplication(ctx context.Context, val string) context.Context {
	return context.WithValue(ctx, applicationKey, val)
}

// GetApplication gets the application value in ctx
func GetApplication(ctx context.Context) string {
	value := ctx.Value(applicationKey)
	if cast, ok := value.(string); ok {
		return cast
	}
	panic("this should not happen, either the original context is not passed or WithApplication is not called as it should")
}
