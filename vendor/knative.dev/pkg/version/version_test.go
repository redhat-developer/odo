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

package version

import (
	"errors"
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/version"
)

type testVersioner struct {
	version string
	err     error
}

func (t *testVersioner) ServerVersion() (*version.Info, error) {
	return &version.Info{GitVersion: t.version}, t.err
}

func TestVersionCheck(t *testing.T) {
	tests := []struct {
		name            string
		actualVersion   *testVersioner
		versionOverride string
		wantError       bool
	}{{
		name:          "greater version (patch)",
		actualVersion: &testVersioner{version: "v1.16.2"},
	}, {
		name:          "greater version (minor)",
		actualVersion: &testVersioner{version: "v1.17.0"},
	}, {
		name:          "same version",
		actualVersion: &testVersioner{version: "v1.16.0"},
	}, {
		name:          "same version with build",
		actualVersion: &testVersioner{version: "v1.16.0+k3s.1"},
	}, {
		name:          "smaller version",
		actualVersion: &testVersioner{version: "v1.14.3"},
		wantError:     true,
	}, {
		name:          "error while fetching",
		actualVersion: &testVersioner{err: errors.New("random error")},
		wantError:     true,
	}, {
		name:            "smaller version with overridden version",
		versionOverride: "v1.13.0",
		actualVersion:   &testVersioner{version: "v1.13.3"},
	}, {
		name:          "unparseable actual version",
		actualVersion: &testVersioner{version: "v1.13.foo"},
		wantError:     true,
	}, {
		name:            "unparseable override version",
		versionOverride: "v1.13.foo",
		actualVersion:   &testVersioner{version: "v1.13.3"},
		wantError:       true,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			os.Setenv(KubernetesMinVersionKey, test.versionOverride)
			defer os.Setenv(KubernetesMinVersionKey, "")

			err := CheckMinimumVersion(test.actualVersion)
			if err == nil && test.wantError {
				t.Errorf("Expected an error for minimum: %q, actual: %v", getMinimumVersion(), test.actualVersion)
			}

			if err != nil && !test.wantError {
				t.Errorf("Expected no error but got %v for minimum: %q, actual: %v", err, getMinimumVersion(), test.actualVersion)
			}
		})
	}
}
