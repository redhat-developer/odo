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

package webhook

import (
	"os"
	"testing"
)

const (
	testDefaultPort      = 8888
	testMissingInputName = "MissingInput" // portEnvKey is not found.
)

type portTest struct {
	name      string
	in        string
	want      int
	wantPanic bool
}

func TestPort(t *testing.T) {
	tests := []portTest{{
		name: testMissingInputName,
		want: testDefaultPort,
	}, {
		name: "EmptyInput",
		in:   "",
		want: testDefaultPort,
	}, {
		name:      "InvalidInputNonNumeric",
		in:        "invalid",
		wantPanic: true,
	}, {
		name:      "InvalidInputTrailingSpace",
		in:        "8443 ",
		wantPanic: true,
	}, {
		name:      "InvalidInputZero",
		in:        "0",
		wantPanic: true,
	}, {
		name: "ValidInput",
		in:   "443",
		want: 443,
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// portEnvKey is unset when testing missing input.
			if tc.name != testMissingInputName {
				os.Setenv(portEnvKey, tc.in)
			}

			defer func() {
				if r := recover(); r == nil && tc.wantPanic {
					t.Error("Did not panic")
				} else if r != nil && !tc.wantPanic {
					t.Error("Got unexpected panic")
				}
				os.Unsetenv(portEnvKey)
			}()

			if got := PortFromEnv(testDefaultPort); got != tc.want {
				t.Errorf("PortFromEnv = %d, want: %d", got, tc.want)
			}
		})
	}
}
