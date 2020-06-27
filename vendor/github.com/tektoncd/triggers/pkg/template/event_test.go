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

package template

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	"github.com/tektoncd/triggers/test"
	bldr "github.com/tektoncd/triggers/test/builder"
	"k8s.io/apimachinery/pkg/runtime"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
)

const (
	ns = "namespace"
)

// toString returns a string representation of a json
func toString(rawMessages []json.RawMessage) []string {
	stringMessages := make([]string, len(rawMessages))
	for i := range rawMessages {
		stringMessages[i] = string(rawMessages[i])
	}
	return stringMessages
}

func TestApplyEventValuesToParams(t *testing.T) {
	var objects = `{"a":"v","c":{"d":"e"},"empty": "","null": null, "number": 42}`
	var arrays = `[{"a": "b"}, {"c": "d"}, {"e": "f"}]`
	tests := []struct {
		name   string
		params []triggersv1.Param
		body   []byte
		header http.Header
		want   []triggersv1.Param
	}{{
		name:   "header with single values",
		params: []triggersv1.Param{bldr.Param("foo", "$(header)")},
		header: map[string][]string{
			"Header-One": {"val1", "val2"},
		},
		want: []triggersv1.Param{bldr.Param("foo", `{"Header-One":"val1,val2"}`)},
	}, {
		name:   "header keys miss-match case",
		params: []triggersv1.Param{bldr.Param("foo", "$(header.header-one)")},
		header: map[string][]string{
			"Header-One": {"val1"},
		},
		want: []triggersv1.Param{bldr.Param("foo", "val1")},
	}, {
		name:   "header keys match case",
		params: []triggersv1.Param{bldr.Param("foo", "$(header.Header-One)")},
		header: map[string][]string{
			"Header-One": {"val1"},
		},
		want: []triggersv1.Param{bldr.Param("foo", "val1")},
	}, {
		name:   "headers - multiple values joined by comma",
		params: []triggersv1.Param{bldr.Param("foo", "$(header.header-one)")},
		header: map[string][]string{
			"Header-One": {"val1", "val2"},
		},
		want: []triggersv1.Param{bldr.Param("foo", "val1,val2")},
	}, {
		name:   "header values",
		params: []triggersv1.Param{bldr.Param("foo", "$(header)")},
		header: map[string][]string{
			"Header-One": {"val1", "val2"},
		},
		want: []triggersv1.Param{bldr.Param("foo", `{"Header-One":"val1,val2"}`)},
	}, {
		name:   "no body",
		params: []triggersv1.Param{bldr.Param("foo", "$(body)")},
		body:   []byte{},
		want:   []triggersv1.Param{bldr.Param("foo", "null")},
	}, {
		name:   "empty body",
		params: []triggersv1.Param{bldr.Param("foo", "$(body)")},
		body:   json.RawMessage(`{}`),
		want:   []triggersv1.Param{bldr.Param("foo", "{}")},
	}, {
		name:   "entire body",
		params: []triggersv1.Param{bldr.Param("foo", "$(body)")},
		body:   json.RawMessage(objects),
		want:   []triggersv1.Param{bldr.Param("foo", strings.ReplaceAll(objects, " ", ""))},
	}, {
		name:   "entire array body",
		params: []triggersv1.Param{bldr.Param("foo", "$(body)")},
		body:   json.RawMessage(arrays),
		want:   []triggersv1.Param{bldr.Param("foo", strings.ReplaceAll(arrays, " ", ""))},
	}, {
		name:   "array key",
		params: []triggersv1.Param{bldr.Param("foo", "$(body.a[1])")},
		body:   json.RawMessage(`{"a": [{"k": 1}, {"k": 2}, {"k": 3}]}`),
		want:   []triggersv1.Param{bldr.Param("foo", `{"k":2}`)},
	}, {
		name:   "array last key",
		params: []triggersv1.Param{bldr.Param("foo", "$(body.a[-1:])")},
		body:   json.RawMessage(`{"a": [{"k": 1}, {"k": 2}, {"k": 3}]}`),
		want:   []triggersv1.Param{bldr.Param("foo", `{"k":3}`)},
	}, {
		name:   "body - key with string val",
		params: []triggersv1.Param{bldr.Param("foo", "$(body.a)")},
		body:   json.RawMessage(objects),
		want:   []triggersv1.Param{bldr.Param("foo", "v")},
	}, {
		name:   "body - key with object val",
		params: []triggersv1.Param{bldr.Param("foo", "$(body.c)")},
		body:   json.RawMessage(objects),
		want:   []triggersv1.Param{bldr.Param("foo", `{"d":"e"}`)},
	}, {
		name:   "body with special chars",
		params: []triggersv1.Param{bldr.Param("foo", "$(body)")},
		body:   json.RawMessage(`{"a": "v\r\n烈"}`),
		want:   []triggersv1.Param{bldr.Param("foo", `{"a":"v\r\n烈"}`)},
	}, {
		name:   "param contains multiple JSONPath expressions",
		params: []triggersv1.Param{bldr.Param("foo", "$(body.a): $(body.b)")},
		body:   json.RawMessage(`{"a": "val1", "b": "val2"}`),
		want:   []triggersv1.Param{bldr.Param("foo", `val1: val2`)},
	}, {
		name:   "param contains both static values and JSONPath expressions",
		params: []triggersv1.Param{bldr.Param("foo", "body.a is: $(body.a)")},
		body:   json.RawMessage(`{"a": "val1"}`),
		want:   []triggersv1.Param{bldr.Param("foo", `body.a is: val1`)},
	}, {
		name: "multiple params",
		params: []triggersv1.Param{
			bldr.Param("foo", "$(body.a)"),
			bldr.Param("bar", "$(header.header-1)"),
		},
		body: json.RawMessage(`{"a": "val1"}`),
		header: map[string][]string{
			"Header-1": {"val2"},
		},
		want: []triggersv1.Param{
			bldr.Param("foo", `val1`),
			bldr.Param("bar", `val2`),
		},
	}, {
		name:   "Array filters",
		body:   json.RawMessage(`{"child":[{"a": "b", "w": "1"}, {"a": "c", "w": "2"}, {"a": "d", "w": "3"}]}`),
		params: []triggersv1.Param{bldr.Param("a", "$(body.child[?(@.a == 'd')].w)")},
		want:   []triggersv1.Param{bldr.Param("a", "3")},
	}, {
		name:   "filters + multiple JSONPath expressions",
		body:   json.RawMessage(`{"child":[{"a": "b", "w": "1"}, {"a": "c", "w": "2"}, {"a": "d", "w": "3"}]}`),
		params: []triggersv1.Param{bldr.Param("a", "$(body.child[?(@.a == 'd')].w) : $(body.child[0].a)")},
		want:   []triggersv1.Param{bldr.Param("a", "3 : b")},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyEventValuesToParams(tt.params, tt.body, tt.header)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.SortSlices(test.CompareParams)); diff != "" {
				t.Errorf("-want/+got: %s", diff)
			}
		})
	}
}

func TestApplyEventValuesToParams_Error(t *testing.T) {
	tests := []struct {
		name   string
		params []triggersv1.Param
		body   []byte
		header http.Header
	}{{
		name:   "missing key",
		params: []triggersv1.Param{bldr.Param("foo", "$(body.missing)")},
		body:   json.RawMessage(`{}`),
	}, {
		name:   "non JSON body",
		params: []triggersv1.Param{bldr.Param("foo", "$(body)")},
		body:   json.RawMessage(`{blahblah}`),
	}, {
		name:   "invalid expression(s)",
		params: []triggersv1.Param{bldr.Param("foo", "$(body.[0])")},
		body:   json.RawMessage(`["a", "b"]`),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyEventValuesToParams(tt.params, tt.body, tt.header)
			if err == nil {
				t.Errorf("did not get expected error - got: %v", got)
			}
		})
	}
}

func TestResolveParams(t *testing.T) {
	tests := []struct {
		name            string
		bindings        []*triggersv1.TriggerBinding
		clusterBindings []*triggersv1.ClusterTriggerBinding
		body            []byte
		template        *triggersv1.TriggerTemplate
		want            []triggersv1.Param
	}{{
		name: "multiple bindings get merged",
		// Two bindings each with a single param
		bindings: []*triggersv1.TriggerBinding{
			bldr.TriggerBinding("b1", ns, bldr.TriggerBindingSpec(
				bldr.TriggerBindingParam("p1", "val1"))),
			bldr.TriggerBinding("b2", ns, bldr.TriggerBindingSpec(
				bldr.TriggerBindingParam("p2", "val2"))),
		},
		template:        bldr.TriggerTemplate("tt", ns),
		clusterBindings: []*triggersv1.ClusterTriggerBinding{},
		want: []triggersv1.Param{
			bldr.Param("p1", "val1"),
			bldr.Param("p2", "val2"),
		},
	}, {
		name: "multiple type bindings get merged",
		// Two bindings each with a single param
		bindings: []*triggersv1.TriggerBinding{
			bldr.TriggerBinding("b1", ns, bldr.TriggerBindingSpec(
				bldr.TriggerBindingParam("p1", "val1"))),
		},
		template: bldr.TriggerTemplate("tt", ns),
		clusterBindings: []*triggersv1.ClusterTriggerBinding{
			bldr.ClusterTriggerBinding("b2", bldr.ClusterTriggerBindingSpec(
				bldr.TriggerBindingParam("p2", "val2"))),
		},
		want: []triggersv1.Param{
			bldr.Param("p1", "val1"),
			bldr.Param("p2", "val2"),
		},
	}, {
		name: "add default values for params with missing values",
		bindings: []*triggersv1.TriggerBinding{
			bldr.TriggerBinding("b1", ns, bldr.TriggerBindingSpec(
				bldr.TriggerBindingParam("p1", "val1"))),
		},
		template: bldr.TriggerTemplate("tt-name", ns,
			bldr.TriggerTemplateSpec(
				bldr.TriggerTemplateParam("p2", "", "defaultVal"),
			),
		),
		clusterBindings: []*triggersv1.ClusterTriggerBinding{},
		want: []triggersv1.Param{
			bldr.Param("p1", "val1"),
			bldr.Param("p2", "defaultVal"),
		},
	}, {
		name: "default values do not override event values",
		bindings: []*triggersv1.TriggerBinding{
			bldr.TriggerBinding("b1", ns, bldr.TriggerBindingSpec(
				bldr.TriggerBindingParam("p1", "val1"))),
		},
		template: bldr.TriggerTemplate("tt-name", ns,
			bldr.TriggerTemplateSpec(
				bldr.TriggerTemplateParam("p1", "", "defaultVal"),
			),
		),
		clusterBindings: []*triggersv1.ClusterTriggerBinding{},
		want: []triggersv1.Param{
			bldr.Param("p1", "val1"),
		},
	}, {
		name: "combination of static values and JSONPath expressions",
		body: json.RawMessage(`{"foo": "fooValue", "bar": "barValue"}`),
		bindings: []*triggersv1.TriggerBinding{
			bldr.TriggerBinding("tb", "ns", bldr.TriggerBindingSpec(
				bldr.TriggerBindingParam("p1", "Event values are - foo: $(body.foo); bar: $(body.bar)"),
			)),
		},
		template:        bldr.TriggerTemplate("tt", ns),
		clusterBindings: []*triggersv1.ClusterTriggerBinding{},
		want: []triggersv1.Param{
			bldr.Param("p1", "Event values are - foo: fooValue; bar: barValue"),
		},
	}, {
		name: "values with newlines",
		body: json.RawMessage(`{"foo": "bar\r\nbaz"}`),
		template: bldr.TriggerTemplate("tt-name", "",
			bldr.TriggerTemplateSpec(
				bldr.TriggerTemplateParam("param1", "", ""),
				bldr.TriggerTemplateParam("param2", "", ""),
			),
		),
		bindings: []*triggersv1.TriggerBinding{
			bldr.TriggerBinding("tb", "namespace",
				bldr.TriggerBindingSpec(
					bldr.TriggerBindingParam("param1", "qux"),
					bldr.TriggerBindingParam("param2", "$(body.foo)"),
				),
			),
		},
		clusterBindings: []*triggersv1.ClusterTriggerBinding{},
		want: []triggersv1.Param{
			bldr.Param("param1", "qux"),
			bldr.Param("param2", "bar\\r\\nbaz"),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := ResolvedTrigger{
				TriggerBindings:        tt.bindings,
				ClusterTriggerBindings: tt.clusterBindings,
				TriggerTemplate:        tt.template,
			}
			params, err := ResolveParams(rt, tt.body, map[string][]string{})
			if err != nil {
				t.Fatalf("ResolveParams() returned unexpected error: %s", err)
			}
			if diff := cmp.Diff(tt.want, params, cmpopts.SortSlices(test.CompareParams)); diff != "" {
				t.Errorf("didn't get expected params -want + got: %s", diff)
			}
		})
	}
}

func TestResolveParams_Error(t *testing.T) {
	tests := []struct {
		name            string
		body            []byte
		params          []triggersv1.ParamSpec
		bindings        []*triggersv1.TriggerBinding
		clusterBindings []*triggersv1.ClusterTriggerBinding
	}{{
		name:            "multiple bindings with same name",
		clusterBindings: []*triggersv1.ClusterTriggerBinding{},
		bindings: []*triggersv1.TriggerBinding{
			bldr.TriggerBinding("b1", ns, bldr.TriggerBindingSpec(
				bldr.TriggerBindingParam("p1", "val1"))),
			bldr.TriggerBinding("b2", ns, bldr.TriggerBindingSpec(
				bldr.TriggerBindingParam("p1", "val2"))),
		},
	}, {
		name:            "invalid body",
		clusterBindings: []*triggersv1.ClusterTriggerBinding{},
		bindings: []*triggersv1.TriggerBinding{
			bldr.TriggerBinding("b1", ns, bldr.TriggerBindingSpec(
				bldr.TriggerBindingParam("p1", "val1"))),
		},
		body: json.RawMessage(`{`),
	}, {
		name:            "invalid expression",
		clusterBindings: []*triggersv1.ClusterTriggerBinding{},
		bindings: []*triggersv1.TriggerBinding{
			bldr.TriggerBinding("b1", ns, bldr.TriggerBindingSpec(
				bldr.TriggerBindingParam("p1", "$(header.[)"))),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := ResolvedTrigger{
				TriggerBindings:        tt.bindings,
				ClusterTriggerBindings: tt.clusterBindings,
			}
			params, err := ResolveParams(rt, tt.body, map[string][]string{})
			if err == nil {
				t.Errorf("did not get expected error - got: %v", params)
			}
		})
	}
}

func TestResolveResources(t *testing.T) {
	tests := []struct {
		name     string
		template *triggersv1.TriggerTemplate
		params   []triggersv1.Param
		want     []json.RawMessage
	}{{
		name: "replace single values in templates",
		template: bldr.TriggerTemplate("tt", ns, bldr.TriggerTemplateSpec(
			bldr.TriggerTemplateParam("p1", "desc", ""),
			bldr.TriggerTemplateParam("p2", "desc", ""),
			bldr.TriggerResourceTemplate(runtime.RawExtension{Raw: []byte(`{"rt1": "$(tt.params.p1)-$(tt.params.p2)"}`)}),
			bldr.TriggerResourceTemplate(runtime.RawExtension{Raw: []byte(`{"rt2": "$(tt.params.p1)-$(tt.params.p2)"}`)}),
		)),
		params: []triggersv1.Param{
			bldr.Param("p1", "val1"),
			bldr.Param("p2", "42"),
		},
		want: []json.RawMessage{
			json.RawMessage(`{"rt1": "val1-42"}`),
			json.RawMessage(`{"rt2": "val1-42"}`),
		},
	}, {
		name: "replace JSON string in templates",
		template: bldr.TriggerTemplate("tt", ns, bldr.TriggerTemplateSpec(
			bldr.TriggerTemplateParam("p1", "desc", ""),
			bldr.TriggerResourceTemplate(runtime.RawExtension{Raw: []byte(`{"rt1": "$(tt.params.p1)"}`)}),
		)),
		params: []triggersv1.Param{
			bldr.Param("p1", `{"a": "b"}`),
		},
		want: []json.RawMessage{
			// json objects get inserted as a valid JSON string
			json.RawMessage(`{"rt1": "{\"a\": \"b\"}"}`),
		},
	}, {
		name: "replace JSON string with special chars in templates",
		template: bldr.TriggerTemplate("tt", ns, bldr.TriggerTemplateSpec(
			bldr.TriggerTemplateParam("p1", "desc", ""),
			bldr.TriggerResourceTemplate(runtime.RawExtension{Raw: []byte(`{"rt1": "$(tt.params.p1)"}`)}),
		)),
		params: []triggersv1.Param{
			bldr.Param("p1", `{"a": "v\\r\\n烈"}`),
		},
		want: []json.RawMessage{
			json.RawMessage(`{"rt1": "{\"a\": \"v\\r\\n烈\"}"}`),
		},
	}, {
		name: "$(uid) gets replaced with a string",
		template: bldr.TriggerTemplate("tt", ns, bldr.TriggerTemplateSpec(
			bldr.TriggerResourceTemplate(runtime.RawExtension{Raw: []byte(`{"rt1": "$(uid)"}`)}),
		)),
		want: []json.RawMessage{
			json.RawMessage(`{"rt1": "cbhtc"}`),
		},
	}, {
		name: "uid replacement is consistent across multiple templates",
		template: bldr.TriggerTemplate("tt", ns, bldr.TriggerTemplateSpec(
			bldr.TriggerResourceTemplate(runtime.RawExtension{Raw: []byte(`{"rt1": "$(uid)"}`)}),
			bldr.TriggerResourceTemplate(runtime.RawExtension{Raw: []byte(`{"rt2": "$(uid)"}`)}),
		)),
		want: []json.RawMessage{
			json.RawMessage(`{"rt1": "cbhtc"}`),
			json.RawMessage(`{"rt2": "cbhtc"}`),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Seeded for UID() to return "cbhtc"
			utilrand.Seed(0)
			got := ResolveResources(tt.template, tt.params)
			// Use toString so that it is easy to compare the json.RawMessage diffs
			if diff := cmp.Diff(toString(tt.want), toString(got)); diff != "" {
				t.Errorf("didn't get expected resource template -want + got: %s", diff)
			}
		})
	}
}
