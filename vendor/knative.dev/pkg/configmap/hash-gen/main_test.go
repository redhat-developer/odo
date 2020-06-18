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

package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestProcess(t *testing.T) {
	for _, test := range []string{"add", "update", "nothing"} {
		t.Run(test, func(t *testing.T) {
			in, err := ioutil.ReadFile(path.Join("testdata", test+".yaml"))
			if err != nil {
				t.Fatalf("Failed to load test fixture: %v", err)
			}

			got, err := process(in)
			if err != nil {
				t.Fatalf("Expected no error but got %v", err)
			}

			want, err := ioutil.ReadFile(path.Join("testdata", test+"_want.yaml"))
			if err != nil && !os.IsNotExist(err) {
				t.Fatalf("Failed to load test fixture: %v", err)
			}

			if diff := cmp.Diff(string(want), string(got)); diff != "" {
				t.Fatal("process (-want, +got) =", diff)
			}
		})
	}
}
