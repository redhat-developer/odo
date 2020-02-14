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

package gke

import (
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/api/option"
)

const credEnvKey = "GOOGLE_APPLICATION_CREDENTIALS"

// func NewSDKClient(opts ...option.ClientOption) (SDKOperations, error) {
func TestNewSDKClient(t *testing.T) {
	pwd, _ := os.Getwd()
	if err := os.Setenv(credEnvKey, filepath.Join(pwd, "fake/credentials.json")); err != nil {
		t.Errorf("Failed to set %s to fake/credentials.json: %v", credEnvKey, err)
	}
	defer os.Unsetenv(credEnvKey)

	datas := []struct {
		req option.ClientOption
	}{{
		// No options.
		nil,
	}, {
		// One option.
		option.WithAPIKey("AIza..."),
	}}
	for _, data := range datas {
		var client SDKOperations
		var err error
		if data.req == nil {
			client, err = NewSDKClient()
		} else {
			client, err = NewSDKClient(data.req)
		}

		if err != nil {
			t.Errorf("Expected no error from request '%v', but got '%v'", data.req, err)
		}
		if client == nil {
			t.Error("Expected a valid client, but got nil")
		}
	}
}
