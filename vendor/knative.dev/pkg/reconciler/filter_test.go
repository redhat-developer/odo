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

package reconciler

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	keyToFilter       = "testKey"
	valueToFilter     = "testVal"
	nameToFilter      = "testName"
	namespaceToFilter = "testSpace"
)

func pod(namespace, name string, annos, labels map[string]string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   namespace,
			Name:        name,
			Annotations: annos,
			Labels:      labels,
		},
	}
}

func podWithLabels(labels map[string]string) *v1.Pod {
	return pod(namespaceToFilter, nameToFilter, nil, labels)
}

func podWithAnnotations(annos map[string]string) *v1.Pod {
	return pod(namespaceToFilter, nameToFilter, annos, nil)
}

func podWithName(name string) *v1.Pod {
	return pod(namespaceToFilter, name, nil, nil)
}

func podWithNamespace(namespace string) *v1.Pod {
	return pod(namespace, nameToFilter, nil, nil)
}

type params struct {
	name       string
	allowUnset bool
	in         interface{}
	want       bool
}

func TestAnnotationFilter(t *testing.T) {
	tests := []params{{
		name: "non kubernetes object",
		in:   struct{}{},
		want: false,
	}, {
		name: "empty annotations",
		in:   podWithAnnotations(nil),
		want: false,
	}, {
		name:       "empty annotations, allow unset",
		allowUnset: true,
		in:         podWithAnnotations(nil),
		want:       true,
	}, {
		name: "other annotations",
		in:   podWithAnnotations(map[string]string{"anotherKey": "anotherValue"}),
		want: false,
	}, {
		name:       "other annotations, allow unset",
		allowUnset: true,
		in:         podWithAnnotations(map[string]string{"anotherKey": "anotherValue"}),
		want:       true,
	}, {
		name: "matching key, value mismatch",
		in:   podWithAnnotations(map[string]string{keyToFilter: "testVal2"}),
		want: false,
	}, {
		name:       "matching key, value mismatch, allow unset",
		allowUnset: true,
		in:         podWithAnnotations(map[string]string{keyToFilter: "testVal2"}),
		want:       false,
	}, {
		name: "match",
		in:   podWithAnnotations(map[string]string{keyToFilter: valueToFilter}),
		want: true,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter := AnnotationFilterFunc(keyToFilter, valueToFilter, test.allowUnset)
			got := filter(test.in)
			if got != test.want {
				t.Errorf("AnnotationFilterFunc() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestLabelExistsFilterFunc(t *testing.T) {
	ti := []params{{
		name: "label exists",
		in:   podWithLabels(map[string]string{keyToFilter: valueToFilter}),
		want: true,
	}, {
		name: "empty labels",
		in:   podWithLabels(map[string]string{}),
		want: false,
	}, {
		name: "non-empty map, the required label doesn't exist",
		in:   podWithLabels(map[string]string{"randomLabel": ""}),
		want: false,
	}, {
		name: "non kubernetes object",
		in:   struct{}{},
		want: false,
	}}

	for _, test := range ti {
		t.Run(test.name, func(t *testing.T) {
			filter := LabelExistsFilterFunc(keyToFilter)
			got := filter(test.in)
			if got != test.want {
				t.Errorf("LabelExistsFilterFunc() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestLabelFilterFunc(t *testing.T) {
	ti := []params{{
		name:       "label matches no unset",
		in:         podWithLabels(map[string]string{keyToFilter: valueToFilter}),
		allowUnset: false,
		want:       true,
	}, {
		name:       "label matches with unset",
		in:         podWithLabels(map[string]string{keyToFilter: valueToFilter}),
		allowUnset: true,
		want:       true,
	}, {
		name:       "label mismatch no unset",
		in:         podWithLabels(map[string]string{keyToFilter: "otherval"}),
		allowUnset: false,
		want:       false,
	}, {
		name:       "label mismatch with unset",
		in:         podWithLabels(map[string]string{keyToFilter: "otherval"}),
		allowUnset: true,
		want:       false,
	}, {
		name:       "label missing no unset",
		in:         podWithLabels(map[string]string{}),
		allowUnset: false,
		want:       false,
	}, {
		name:       "label missing with unset",
		in:         podWithLabels(map[string]string{}),
		allowUnset: true,
		want:       true,
	}, {
		name:       "nil labels no unset",
		in:         podWithLabels(nil),
		allowUnset: false,
		want:       false,
	}, {
		name:       "nil labels with unset",
		in:         podWithLabels(nil),
		allowUnset: true,
		want:       true,
	}, {
		name: "non kubernetes object",
		in:   struct{}{},
		want: false,
	}}

	for _, test := range ti {
		t.Run(test.name, func(t *testing.T) {
			filter := LabelFilterFunc(keyToFilter, valueToFilter, test.allowUnset)
			got := filter(test.in)
			if got != test.want {
				t.Errorf("LabelFilterFunc() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestNameFilterFunc(t *testing.T) {
	ti := []params{{
		name: "name match",
		in:   podWithName(nameToFilter),
		want: true,
	}, {
		name: "name mismatch",
		in:   podWithName("bogus"),
		want: false,
	}, {
		name: "non kubernetes object",
		in:   struct{}{},
		want: false,
	}}

	for _, test := range ti {
		t.Run(test.name, func(t *testing.T) {
			filter := NameFilterFunc(nameToFilter)
			got := filter(test.in)
			if got != test.want {
				t.Errorf("NameFilterFunc() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestNamespaceFilterFunc(t *testing.T) {
	ti := []params{{
		name: "namespace match",
		in:   podWithNamespace(namespaceToFilter),
		want: true,
	}, {
		name: "namespace mismatch",
		in:   podWithNamespace("bogus"),
		want: false,
	}, {
		name: "non kubernetes object",
		in:   struct{}{},
		want: false,
	}}

	for _, test := range ti {
		t.Run(test.name, func(t *testing.T) {
			filter := NamespaceFilterFunc(namespaceToFilter)
			got := filter(test.in)
			if got != test.want {
				t.Errorf("NamespaceFilterFunc() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestChainFilterFuncs(t *testing.T) {
	tc := []struct {
		name  string
		chain []bool
		want  bool
	}{{
		name:  "single true",
		chain: []bool{true},
		want:  true,
	}, {
		name:  "single false",
		chain: []bool{false},
		want:  false,
	}, {
		name:  "second false",
		chain: []bool{true, false},
		want:  false,
	}, {
		name:  "multi true",
		chain: []bool{true, true},
		want:  true,
	}}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			filters := make([]func(interface{}) bool, len(test.chain))
			for i, chainVal := range test.chain {
				filters[i] = func(interface{}) bool {
					return chainVal
				}
			}
			filter := ChainFilterFuncs(filters...)
			got := filter(nil)
			if got != test.want {
				t.Errorf("ChainFilterFuncs() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestNotFilter(t *testing.T) {
	odd := func(o interface{}) bool {
		// Return true if odd.
		return (o.(int))&1 == 1
	}
	if got, want := Not(odd)(1), false; got != want {
		t.Errorf("Odd input = %v, want: %v", got, want)
	}
	if got, want := Not(odd)(2), true; got != want {
		t.Errorf("Odd input = %v, want: %v", got, want)
	}
}
