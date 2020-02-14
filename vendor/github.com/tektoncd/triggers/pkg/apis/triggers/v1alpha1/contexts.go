package v1alpha1

import "context"

// upgradeViaDefaultingKey is used as the key in a context.Context.
// This variable doesn't really matter, so it can be a total random name.
// Setting this key indicates that default values for a resource should be
// updated to new values. This is used to ensure non breaking updates when
// a default value of a resource changes or when a field is removed.
type upgradeViaDefaultingKey struct{}

// WithUpgradeViaDefaulting sets the upgradeViaDefaultingKey on the context
// indicating that default values for a resource should be updated to new values.
func WithUpgradeViaDefaulting(ctx context.Context) context.Context {
	return context.WithValue(ctx, upgradeViaDefaultingKey{}, struct{}{})
}

// IsUpgradeViaDefaulting checks if the upgradeViaDefaultingKey is set on
// the context.
func IsUpgradeViaDefaulting(ctx context.Context) bool {
	return ctx.Value(upgradeViaDefaultingKey{}) != nil
}
