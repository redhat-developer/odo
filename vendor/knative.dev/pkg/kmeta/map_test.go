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

package kmeta

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnion(t *testing.T) {
	tests := []struct {
		name string
		in   []map[string]string
		want map[string]string
	}{{
		name: "nothing at all",
		want: map[string]string{},
	}, {
		name: "empty in",
		in:   []map[string]string{},
		want: map[string]string{},
	}, {
		name: "nil in list",
		in:   []map[string]string{nil},
		want: map[string]string{},
	}, {
		name: "empty in recursive",
		in:   []map[string]string{{}},
		want: map[string]string{},
	}, {
		name: "nil and something",
		in:   []map[string]string{nil, {"wish": "you", "were": "here"}},
		want: map[string]string{"wish": "you", "were": "here"},
	}, {
		name: "something and nil",
		in:   []map[string]string{{"the-dark": "side"}},
		want: map[string]string{"the-dark": "side"},
	}, {
		name: "2 in",
		in:   []map[string]string{{"another": "brick"}, {"in": "the-wall"}},
		want: map[string]string{"in": "the-wall", "another": "brick"},
	}, {
		name: "2 in, latter wins",
		in:   []map[string]string{{"another": "brick", "in": "the-wall-pt-I"}, {"in": "the-wall-pt-II"}},
		want: map[string]string{"in": "the-wall-pt-II", "another": "brick"},
	}, {
		name: "3 in, latter wins",
		in: []map[string]string{
			{"another": "brick", "in": "the-wall-pt-I"},
			{"in": "the-wall-pt-II"},
			{"in": "the-wall-pt-III", "hey": "you"},
		},
		want: map[string]string{"in": "the-wall-pt-III", "another": "brick", "hey": "you"},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := UnionMaps(test.in...)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("MakeLabels (-want, +got) = %v", diff)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	tests := []struct {
		name   string
		in     map[string]string
		filter func(string) bool
		want   map[string]string
	}{{
		name: "nil in",
		want: map[string]string{},
	}, {
		name: "empty in",
		in:   map[string]string{},
		want: map[string]string{},
	}, {
		name:   "no in, with filter",
		in:     map[string]string{},
		filter: func(string) bool { return false },
		want:   map[string]string{},
	}, {
		name: "pass through",
		in:   map[string]string{"the-dark": "side"},
		want: map[string]string{"the-dark": "side"},
	}, {
		name:   "filter all",
		in:     map[string]string{"the-dark": "side", "of-there": "moon"},
		filter: func(string) bool { return true },
		want:   map[string]string{},
	}, {
		name:   "all together now",
		in:     map[string]string{"another": "brick", "in": "the-wall"},
		filter: func(s string) bool { return s == "in" },
		want:   map[string]string{"another": "brick"},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := FilterMap(test.in, test.filter)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("MakeAnnotations (-want, +got) = %v", diff)
			}
		})
	}
}

func TestCopy(t *testing.T) {
	tests := []struct {
		name string
		in   map[string]string
		want map[string]string
	}{{
		name: "nil in",
		want: map[string]string{},
	}, {
		name: "empty in",
		in:   map[string]string{},
		want: map[string]string{},
	}, {
		name: "copy",
		in:   map[string]string{"the-dark": "side"},
		want: map[string]string{"the-dark": "side"},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := CopyMap(test.in)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("MakeAnnotations (-want, +got) = %v", diff)
			}
		})
	}
}
