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

import (
	"context"

	"knative.dev/pkg/apis"
)

// Validate TriggerBinding.
func (tb *TriggerBinding) Validate(ctx context.Context) *apis.FieldError {
	return tb.Spec.Validate(ctx)
}

// Validate TriggerBindingSpec.
func (s *TriggerBindingSpec) Validate(ctx context.Context) *apis.FieldError {
	if err := validateParams(s.Params); err != nil {
		return err
	}
	return nil
}

func validateParams(params []Param) *apis.FieldError {
	// Ensure there aren't multiple params with the same name.
	seen := map[string]struct{}{}
	for _, param := range params {
		if _, ok := seen[param.Name]; ok {
			return apis.ErrMultipleOneOf("spec.params")
		}
		seen[param.Name] = struct{}{}
	}
	return nil
}
