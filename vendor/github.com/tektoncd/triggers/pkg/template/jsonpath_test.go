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
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var objects = `{"a":"v\r\n烈","c":{"d":"e"},"empty": "","null": null, "number": 42}`
var arrays = `[{"a": "b"}, {"c": "d"}, {"e": "f"}]`

// Checks that we print JSON strings when the JSONPath selects
// an array or map value and regular values otherwise
func TestParseJSONPath(t *testing.T) {
	var objectBody = fmt.Sprintf(`{"body":%s}`, objects)
	var arrayBody = fmt.Sprintf(`{"body":%s}`, arrays)
	tests := []struct {
		name string
		expr string
		in   string
		want string
	}{{
		name: "objects",
		in:   objectBody,
		expr: "$(body)",
		// TODO: Do we need to escape backslashes for backwards compat?
		want: objects,
	}, {
		name: "array of objects",
		in:   arrayBody,
		expr: "$(body)",
		want: arrays,
	}, {
		name: "array of values",
		in:   `{"body": ["a", "b", "c"]}`,
		expr: "$(body)",
		want: `["a", "b", "c"]`,
	}, {
		name: "string values",
		in:   objectBody,
		expr: "$(body.a)",
		want: "v\\r\\n烈",
	}, {
		name: "empty string",
		in:   objectBody,
		expr: "$(body.empty)",
		want: "",
	}, {
		name: "numbers",
		in:   objectBody,
		expr: "$(body.number)",
		want: "42",
	}, {
		name: "booleans",
		in:   `{"body": {"bool": true}}`,
		expr: "$(body.bool)",
		want: "true",
	}, {
		name: "null values",
		in:   objectBody,
		expr: "$(body.null)",
		want: "null",
	}, {
		name: "multiple results",
		in:   arrayBody,
		expr: "$(body[:2])",
		want: `[{"a": "b"}, {"c": "d"}]`,
	}, {
		name: "multiple results with empty string",
		in:   `{"body":["", "some", "thing"]}`,
		expr: "$(body[:2])",
		want: `["", "some"]`,
	}, {
		name: "multiple results newlines/special chars",
		in:   `{"body":["", "v\r\n烈", "thing"]}`,
		expr: "$(body[:2])",
		want: `["", "v\r\n烈"]`,
	}, {
		name: "multiple results with null",
		in:   `{"body":["", null, "thing"]}`,
		expr: "$(body[:2])",
		want: `["", null]`,
	}, {
		name: "Array filters",
		in:   `{"body":{"child":[{"a": "b", "w": "1"}, {"a": "c", "w": "2"}, {"a": "d", "w": "3"}]}}`,
		expr: "$(body.child[?(@.a == 'd')].w)",
		want: "3",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data interface{}
			err := json.Unmarshal([]byte(tt.in), &data)
			if err != nil {
				t.Fatalf("Could not unmarshall body : %q", err)
			}
			got, err := ParseJSONPath(data, tt.expr)
			if err != nil {
				t.Fatalf("ParseJSONPath() error = %v", err)
			}
			if diff := cmp.Diff(strings.Replace(tt.want, " ", "", -1), got); diff != "" {
				t.Errorf("ParseJSONPath() -want,+got: %s", diff)
			}
		})
	}
}

func TestParseJSONPath_Error(t *testing.T) {
	testJSON := `{"body": {"key": "val"}}`
	invalidExprs := []string{
		"$({.hello)",
		"$(+12.3.0)",
		"$([1)",
		"$(body",
		"body)",
		"body",
		"$(body.missing)",
		"$(body.key[0])",
	}
	var data interface{}
	err := json.Unmarshal([]byte(testJSON), &data)
	if err != nil {
		t.Fatalf("Could not unmarshall body : %q", err)
	}

	for _, expr := range invalidExprs {
		t.Run(expr, func(t *testing.T) {
			got, err := ParseJSONPath(data, expr)
			if err == nil {
				t.Errorf("ParseJSONPath() did not return expected error; got = %v", got)
			}
		})
	}
}

func TestTektonJSONPathExpression(t *testing.T) {
	tests := []struct {
		expr string
		want string
	}{
		{"$(metadata.name)", "{.metadata.name}"},
		{"$(.metadata.name)", "{.metadata.name}"},
		{"$({.metadata.name})", "{.metadata.name}"},
		{"$()", ""},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			got, err := TektonJSONPathExpression(tt.expr)
			if err != nil {
				t.Errorf("TektonJSONPathExpression() unexpected error = %v,  got = %v", err, got)
			}
			if got != tt.want {
				t.Errorf("TektonJSONPathExpression() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTektonJSONPathExpression_Error(t *testing.T) {
	tests := []string{
		"{.metadata.name}", // not wrapped in $()
		"",
		"$({asd)",
		"$({)",
		"$({foo.bar)",
		"$(foo.bar})",
		"$({foo.bar}})",
		"$({{foo.bar)",
	}
	for _, expr := range tests {
		t.Run(expr, func(t *testing.T) {
			_, err := TektonJSONPathExpression(expr)
			if err == nil {
				t.Errorf("TektonJSONPathExpression() did not get expected error for expression = %s", expr)
			}
		})
	}
}

func TestRelaxedJSONPathExpression(t *testing.T) {
	tests := []struct {
		expr string
		want string
	}{
		{"metadata.name", "{.metadata.name}"},
		{".metadata.name", "{.metadata.name}"},
		{"{.metadata.name}", "{.metadata.name}"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			got, err := relaxedJSONPathExpression(tt.expr)
			if err != nil {
				t.Errorf("TektonJSONPathExpression() unexpected error = %v,  got = %v", err, got)
			}
			if got != tt.want {
				t.Errorf("TektonJSONPathExpression() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRelaxedJSONPathExpression_Error(t *testing.T) {
	tests := []string{
		"{foo.bar",
		"foo.bar}",
		"{foo.bar}}",
		"{{foo.bar}",
	}
	for _, expr := range tests {
		t.Run(expr, func(t *testing.T) {
			got, err := relaxedJSONPathExpression(expr)
			if err == nil {
				t.Errorf("TektonJSONPathExpression() did not get expected error = %v,  got = %v", err, got)
			}
		})
	}
}

func TestFindTektonExpressions(t *testing.T) {
	tcs := []struct {
		in       string
		want     []string
		original []string
	}{{
		in:       "$(body.blah)",
		want:     []string{"$(body.blah)"},
		original: []string{"$(body.blah)"},
	}, {
		in:       "$(body.blah)-$(header.*)",
		want:     []string{"$(body.blah)", "$(header.*)"},
		original: []string{"$(body.blah)", "$(header.*)"},
	}, {
		in:       "start:$(body.blah)//middle//$(header.one)-end",
		want:     []string{"$(body.blah)", "$(header.One)"},
		original: []string{"$(body.blah)", "$(header.one)"},
	}, {
		in:       "start:$(body.blah)//middle//$(header.One)-end",
		want:     []string{"$(body.blah)", "$(header.One)"},
		original: []string{"$(body.blah)", "$(header.One)"},
	}, {
		in:       "start:$(body.blah)//middle//$(header.ONE-TWO)-end",
		want:     []string{"$(body.blah)", "$(header.One-Two)"},
		original: []string{"$(body.blah)", "$(header.ONE-TWO)"},
	}, {
		in:       "start:$(body.[?(@.a == 'd')])-$(body.another-one)",
		want:     []string{"$(body.[?(@.a == 'd')])", "$(body.another-one)"},
		original: []string{"$(body.[?(@.a == 'd')])", "$(body.another-one)"},
	}, {
		in:       "$(this)-$(not-this",
		want:     []string{"$(this)"},
		original: []string{"$(this)"},
	}, {
		in:       "$body.)",
		want:     []string{},
		original: []string{},
	}, {
		in:       "($(body.blah))-and-$(body.foo)",
		want:     []string{"$(body.blah)", "$(body.foo)"},
		original: []string{"$(body.blah)", "$(body.foo)"},
	}, {
		in:       "(staticvalue)$(body.blah)",
		want:     []string{"$(body.blah)"},
		original: []string{"$(body.blah)"},
	}, {
		in:       "asd)$(asd",
		want:     []string{},
		original: []string{},
	}, {
		in:       "onlystatic",
		want:     []string{},
		original: []string{},
	}, {
		in:       "",
		want:     []string{},
		original: []string{},
	}, {
		in:       "$())))",
		want:     []string{"$()"},
		original: []string{"$()"},
	}, {
		in:       "$($())",
		want:     []string{"$()"},
		original: []string{"$()"},
	}, {
		in:       "$($($(blahblah)))",
		want:     []string{"$(blahblah)"},
		original: []string{"$(blahblah)"},
	}}

	for _, tc := range tcs {
		t.Run(tc.in, func(t *testing.T) {
			results, originals := findTektonExpressions(tc.in)
			if diff := cmp.Diff(tc.want, results); diff != "" {
				t.Fatalf("error -want/+got: %s", diff)
			}
			if diff := cmp.Diff(tc.original, originals); diff != "" {
				t.Fatalf("error -want/+got: %s", diff)
			}
		})
	}
}
