/*
Copyright 2020 The Tekton Authors

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

package cmd

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEvalBinding(t *testing.T) {
	out := new(bytes.Buffer)
	if err := evalBinding(out, "../testdata/triggerbinding.yaml", "../testdata/http.txt"); err != nil {
		t.Fatalf("evalBinding: %v", err)
	}

	want := `[
  {
    "name": "bar",
    "value": "tacocat"
  },
  {
    "name": "foo",
    "value": "body"
  }
]
`
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("-want +got: %s", diff)
	}
}
