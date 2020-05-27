/*
Copyright 2019 The Tekton Authors

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

package builder

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestObjectMeta(t *testing.T) {
	tests := []struct {
		name         string
		normal       *metav1.ObjectMeta
		builderFuncs []ObjectMetaOp
	}{
		{
			name: "One Label",
			normal: &metav1.ObjectMeta{
				Labels: map[string]string{
					"key1": "value1",
				},
			},
			builderFuncs: []ObjectMetaOp{
				Label("key1", "value1"),
			},
		},
		{
			name: "Two Labels",
			normal: &metav1.ObjectMeta{
				Labels: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
			builderFuncs: []ObjectMetaOp{
				Label("key1", "value1"),
				Label("key2", "value2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objectMetaBuilder := &metav1.ObjectMeta{}
			for _, op := range tt.builderFuncs {
				op(objectMetaBuilder)
			}
			if diff := cmp.Diff(tt.normal, objectMetaBuilder); diff != "" {
				t.Errorf("TestObjectMeta(): -want +got: %s", diff)
			}
		})
	}
}

func TestTypeMeta(t *testing.T) {
	typeMeta := new(metav1.TypeMeta)
	TypeMeta("kind", "version")(typeMeta)
	expectedTypeMeta := &metav1.TypeMeta{
		Kind:       "kind",
		APIVersion: "version",
	}
	if d := cmp.Diff(expectedTypeMeta, typeMeta); d != "" {
		t.Fatalf("Pod diff -want, +got:\n%v", d)
	}
}
