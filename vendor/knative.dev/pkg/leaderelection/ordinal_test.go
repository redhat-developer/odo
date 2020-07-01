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

package leaderelection

import (
	"os"
	"testing"
)

const controllerOrdinalEnv = "STATEFUL_CONTROLLER_ORDINAL"

func TestControllerOrdinal(t *testing.T) {
	testCases := []struct {
		testname    string
		podName     string
		wantName    string
		wantOrdinal int
		wantErr     bool
	}{{
		testname: "NotSet",
		wantErr:  true,
	}, {
		testname: "NoHyphen",
		podName:  "as",
		wantErr:  true,
	}, {
		testname: "InvalidOrdinal",
		podName:  "as-invalid",
		wantErr:  true,
	}, {
		testname: "ValidName",
		podName:  "as-0",
		wantName: "as",
	}, {
		testname:    "ValidName",
		podName:     "as-1",
		wantName:    "as",
		wantOrdinal: 1,
	}}

	defer os.Unsetenv(controllerOrdinalEnv)
	for _, tt := range testCases {
		t.Run(tt.testname, func(t *testing.T) {
			if tt.podName != "" {
				if os.Setenv(controllerOrdinalEnv, tt.podName) != nil {
					t.Fatalf("fail to set env var %s=%s", controllerOrdinalEnv, tt.podName)
				}
				os.Setenv("STATEFUL_SERVICE_NAME", "n'importe quoi")
				os.Setenv("STATEFUL_SERVICE_PORT", "1299")
				os.Setenv("STATEFUL_SERVICE_PROTOCOL", "n'importe quoi")
			}

			gotOrdinal, gotOrdinalErr := ControllerOrdinal()
			if (gotOrdinalErr != nil) != tt.wantErr {
				t.Fatalf("Err = %v, wantErr = %v", gotOrdinalErr, tt.wantErr)
			}
			if gotOrdinal != tt.wantOrdinal {
				t.Errorf("ControllerOrdinal() = %d, want = %d", gotOrdinal, tt.wantOrdinal)
			}
		})
	}
}
