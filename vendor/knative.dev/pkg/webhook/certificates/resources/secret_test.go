/*
Copyright 2019 The Knative Authors

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

package resources

import (
	"testing"

	. "knative.dev/pkg/logging/testing"
)

func TestMakeSecret(t *testing.T) {
	ctx := TestContextWithLogger(t)
	secret, err := MakeSecret(ctx, "foo", "ns", "bar")
	if err != nil {
		t.Errorf("MakeSecret() = %v", err)
	}

	for _, key := range []string{ServerKey, ServerCert, CACert} {
		if _, ok := secret.Data[key]; !ok {
			t.Errorf("secret.Data[%q] is missing", key)
		}
	}
}
