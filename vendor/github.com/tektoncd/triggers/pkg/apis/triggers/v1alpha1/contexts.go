/*
Copyright 2019 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
