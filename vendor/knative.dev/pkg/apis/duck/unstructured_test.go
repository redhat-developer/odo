/*
Copyright 2018 The Knative Authors

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

package duck

import (
	"testing"

	"encoding/json"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	. "knative.dev/pkg/testing"
)

func TestFromUnstructuredFooable(t *testing.T) {
	tcs := []struct {
		name      string
		in        json.Marshaler
		want      FooStatus
		wantError error
	}{{
		name: "Works with valid status",
		in: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "test",
				"kind":       "test_kind",
				"name":       "test_name",
				"status": map[string]interface{}{
					"extra": "fields",
					"fooable": map[string]interface{}{
						"field1": "foo",
						"field2": "bar",
					},
				},
			}},
		want: FooStatus{&Fooable{
			Field1: "foo",
			Field2: "bar",
		}},
		wantError: nil,
	}, {
		name: "does not work with missing fooable status",
		in: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "test",
				"kind":       "test_kind",
				"name":       "test_name",
				"status": map[string]interface{}{
					"extra": "fields",
				},
			}},
		want:      FooStatus{},
		wantError: nil,
	}, {
		name:      "empty unstructured",
		in:        &unstructured.Unstructured{},
		want:      FooStatus{},
		wantError: nil,
	}}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(tc.in)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			t.Logf("Marshalled: %s", string(raw))

			got := Foo{}
			if err := FromUnstructured(tc.in, &got); err != tc.wantError {
				t.Fatalf("FromUnstructured() = %v", err)
			}

			if !cmp.Equal(tc.want, got.Status) {
				t.Errorf("ToUnstructured (-want, +got) = %s", cmp.Diff(tc.want, got.Status))
			}
		})
	}
}

func TestToUnstructured(t *testing.T) {
	tests := []struct {
		name      string
		in        OneOfOurs
		want      *unstructured.Unstructured
		wantError error
	}{{
		name: "missing TypeMeta",
		in: &Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name: "blah",
			},
		},
		want: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "pkg.knative.dev/v2",
				"kind":       "Resource",
				"metadata": map[string]interface{}{
					"creationTimestamp": nil,
					"name":              "blah",
				},
				"spec": map[string]interface{}{},
			},
		},
	}, {
		name: "with TypeMeta",
		in: &Resource{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "blah",
			},
		},
		want: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"creationTimestamp": nil,
					"name":              "blah",
				},
				"spec": map[string]interface{}{},
			},
		},
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ToUnstructured(tc.in)
			if err != tc.wantError {
				t.Fatalf("ToUnstructured() = %v", err)
			}

			if !cmp.Equal(tc.want, got) {
				t.Errorf("ToUnstructured (-want, +got) = %s", cmp.Diff(tc.want, got))
			}
		})
	}
}
