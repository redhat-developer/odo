/*
Copyright 2020 The Knative Authors

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

package psbinding

import (
	"context"
	"testing"
)

func TestDefault(t *testing.T) {
	table := []struct {
		name   string
		in     context.Context
		optout bool
	}{{
		name:   "default",
		in:     context.Background(),
		optout: false,
	}, {
		name:   "default",
		in:     WithOptOutSelector(context.Background()),
		optout: true,
	}}

	for _, tc := range table {
		if want, got := tc.optout, HasOptOutSelector(tc.in); want != got {
			t.Errorf("Unexpected optout (-want, +got): %v, %v", want, got)
		}
	}
}
