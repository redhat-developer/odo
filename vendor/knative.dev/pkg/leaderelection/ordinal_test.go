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
	"fmt"
	"os"
	"testing"
)

func TestControllerOrdinal(t *testing.T) {
	testCases := []struct {
		testname string
		podName  string
		want     uint64
		err      error
	}{{
		testname: "NotSet",
		err:      fmt.Errorf("ordinal not found in %s=", controllerOrdinalEnv),
	}, {
		testname: "NoHyphen",
		podName:  "as",
		err:      fmt.Errorf("ordinal not found in %s=as", controllerOrdinalEnv),
	}, {
		testname: "InvalidOrdinal",
		podName:  "as-invalid",
		err:      fmt.Errorf(`strconv.ParseUint: parsing "invalid": invalid syntax`),
	}, {
		testname: "ValidName",
		podName:  "as-0",
	}, {
		testname: "ValidName",
		podName:  "as-1",
		want:     1,
	}}

	defer os.Unsetenv(controllerOrdinalEnv)
	for _, tt := range testCases {
		t.Run(tt.testname, func(t *testing.T) {
			if tt.podName != "" {
				if os.Setenv(controllerOrdinalEnv, tt.podName) != nil {
					t.Fatalf("fail to set env var %s=%s", controllerOrdinalEnv, tt.podName)
				}
			}

			got, gotErr := ControllerOrdinal()
			if tt.err != nil {
				if gotErr == nil || gotErr.Error() != tt.err.Error() {
					t.Errorf("got %v, want = %v, ", gotErr, tt.err)
				}
			} else if gotErr != nil {
				t.Error("ControllerOrdinal() =", gotErr)
			} else if got != tt.want {
				t.Errorf("ControllerOrdinal() = %d, want = %d", got, tt.want)
			}
		})
	}
}
