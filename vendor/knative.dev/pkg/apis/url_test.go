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
package apis

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/api/equality"
)

func TestParseURL(t *testing.T) {
	testCases := map[string]struct {
		t         string
		want      *URL
		wantEmpty bool
		wantErr   bool
	}{
		"empty string": {
			want:      nil,
			wantEmpty: true,
		},
		"invalid format": {
			t:         "💩://error",
			want:      nil,
			wantEmpty: true,
			wantErr:   true,
		},
		"relative": {
			t: "/path/to/something",
			want: &URL{
				Path: "/path/to/something",
			},
		},
		"url": {
			t: "http://path/to/something",
			want: &URL{
				Scheme: "http",
				Host:   "path",
				Path:   "/to/something",
			},
		},
		"simplehttp": {
			t:    "http://foo",
			want: HTTP("foo"),
		},
		"simplehttps": {
			t:    "https://foo",
			want: HTTPS("foo"),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got, err := ParseURL(tc.t)
			if err != nil {
				if !tc.wantErr {
					t.Fatalf("ParseURL() = %v", err)
				}
				return
			} else if tc.wantErr {
				t.Fatalf("ParseURL() = %v, wanted error", got)
			}

			if tc.wantEmpty != got.IsEmpty() {
				t.Errorf("IsEmpty(%v) = %t, wanted %t", got, got.IsEmpty(), tc.wantEmpty)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected object (-want, +got) = %v", diff)
			}
		})
	}
}

func TestJsonMarshalURL(t *testing.T) {
	testCases := map[string]struct {
		t    string
		want []byte
	}{
		"empty": {},
		"empty string": {
			t: "",
		},
		"invalid url": {
			t:    "not a url",
			want: []byte(`"not%20a%20url"`),
		},
		"relative format": {
			t:    "/path/to/something",
			want: []byte(`"/path/to/something"`),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			var got []byte
			tt, err := ParseURL(tc.t)
			if err != nil {
				t.Fatalf("ParseURL() = %v", err)
			}
			if tt != nil {
				got, _ = tt.MarshalJSON()
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Logf("got: %s", string(got))
				t.Errorf("unexpected object (-want, +got) = %v", diff)
			}
		})
	}
}

func TestJsonUnmarshalURL(t *testing.T) {
	testCases := map[string]struct {
		b       []byte
		want    *URL
		wantErr string
	}{
		"empty": {
			wantErr: "unexpected end of JSON input",
		},
		"invalid format": {
			b:       []byte("%"),
			wantErr: "invalid character '%' looking for beginning of value",
		},
		"relative": {
			b: []byte(`"/path/to/something"`),
			want: &URL{
				Path: "/path/to/something",
			},
		},
		"url": {
			b: []byte(`"http://path/to/something"`),
			want: &URL{
				Scheme: "http",
				Host:   "path",
				Path:   "/to/something",
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := &URL{}
			err := got.UnmarshalJSON(tc.b)

			if tc.wantErr != "" || err != nil {
				var gotErr string
				if err != nil {
					gotErr = err.Error()
				}
				if diff := cmp.Diff(tc.wantErr, gotErr); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected object (-want, +got) = %v", diff)
			}
		})
	}
}

func TestJsonMarshalURLAsMember(t *testing.T) {

	type objectType struct {
		URL URL `json:"url,omitempty"`
	}

	testCases := map[string]struct {
		obj     *objectType
		want    []byte
		wantErr string
	}{
		"nil": {
			want: []byte(`null`),
		},
		"empty": {
			obj:  &objectType{},
			want: []byte(`{"url":""}`),
		},
		"relative": {
			obj:  &objectType{URL: URL{Path: "/path/to/something"}},
			want: []byte(`{"url":"/path/to/something"}`),
		},
		"url": {
			obj: &objectType{URL: URL{
				Scheme: "http",
				Host:   "path",
				Path:   "/to/something",
			}},
			want: []byte(`{"url":"http://path/to/something"}`),
		},
		"empty url": {
			obj:  &objectType{URL: URL{}},
			want: []byte(`{"url":""}`),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got, err := json.Marshal(tc.obj)

			if tc.wantErr != "" || err != nil {
				var gotErr string
				if err != nil {
					gotErr = err.Error()
				}
				if diff := cmp.Diff(tc.wantErr, gotErr); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected object (-want, +got) = %v", diff)
				t.Logf("got: %s", string(got))
			}
		})
	}
}

func TestJsonMarshalURLAsPointerMember(t *testing.T) {

	type objectType struct {
		URL *URL `json:"url,omitempty"`
	}

	testCases := map[string]struct {
		obj     *objectType
		want    []byte
		wantErr string
	}{
		"nil": {
			want: []byte(`null`),
		},
		"empty": {
			obj:  &objectType{},
			want: []byte(`{}`),
		},
		"relative": {
			obj:  &objectType{URL: &URL{Path: "/path/to/something"}},
			want: []byte(`{"url":"/path/to/something"}`),
		},
		"url": {
			obj: &objectType{URL: &URL{
				Scheme: "http",
				Host:   "path",
				Path:   "/to/something",
			}},
			want: []byte(`{"url":"http://path/to/something"}`),
		},
		"empty url": {
			obj:  &objectType{URL: &URL{}},
			want: []byte(`{"url":""}`),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got, err := json.Marshal(tc.obj)

			if tc.wantErr != "" || err != nil {
				var gotErr string
				if err != nil {
					gotErr = err.Error()
				}
				if diff := cmp.Diff(tc.wantErr, gotErr); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected object (-want, +got) = %v", diff)
				t.Logf("got: %s", string(got))
			}
		})
	}
}

func TestJsonUnmarshalURLAsMember(t *testing.T) {

	type objectType struct {
		URL URL `json:"url,omitempty"`
	}

	testCases := map[string]struct {
		b       []byte
		want    *objectType
		wantErr string
	}{
		"zero": {
			wantErr: "unexpected end of JSON input",
		},
		"empty": {
			b:    []byte(`{}`),
			want: &objectType{},
		},
		"invalid format": {
			b:       []byte(`{"url":"%"}`),
			wantErr: `invalid URL escape "%"`,
		},
		"relative": {
			b:    []byte(`{"url":"/path/to/something"}`),
			want: &objectType{URL: URL{Path: "/path/to/something"}},
		},
		"url": {
			b: []byte(`{"url":"http://path/to/something"}`),
			want: &objectType{URL: URL{
				Scheme: "http",
				Host:   "path",
				Path:   "/to/something",
			}},
		},
		"empty url": {
			b:    []byte(`{"url":""}`),
			want: &objectType{URL: URL{}},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := &objectType{}
			err := json.Unmarshal(tc.b, got)

			if tc.wantErr != "" || err != nil {
				var gotErr string
				if err != nil {
					gotErr = err.Error()
				}
				if !strings.Contains(gotErr, tc.wantErr) {
					t.Errorf("Error `%s` does not contain wanted string `%s`", gotErr, tc.wantErr)
				}
				return
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Unexpected object (-want, +got) = %v", diff)
			}
		})
	}
}

func TestJsonUnmarshalURLAsMemberPointer(t *testing.T) {

	type objectType struct {
		URL *URL `json:"url,omitempty"`
	}

	testCases := map[string]struct {
		b       []byte
		want    *objectType
		wantErr string
	}{
		"zero": {
			wantErr: "unexpected end of JSON input",
		},
		"empty": {
			b:    []byte(`{}`),
			want: &objectType{},
		},
		"invalid format": {
			b:       []byte(`{"url":"%"}`),
			wantErr: `invalid URL escape "%"`,
		},
		"relative": {
			b:    []byte(`{"url":"/path/to/something"}`),
			want: &objectType{URL: &URL{Path: "/path/to/something"}},
		},
		"url": {
			b: []byte(`{"url":"http://path/to/something"}`),
			want: &objectType{URL: &URL{
				Scheme: "http",
				Host:   "path",
				Path:   "/to/something",
			}},
		},
		"empty url": {
			b:    []byte(`{"url":""}`),
			want: &objectType{URL: &URL{}},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := &objectType{}
			err := json.Unmarshal(tc.b, got)

			if tc.wantErr != "" || err != nil {
				var gotErr string
				if err != nil {
					gotErr = err.Error()
				}
				if !strings.Contains(gotErr, tc.wantErr) {
					t.Errorf("Error `%s` does not contain wanted string `%s`", gotErr, tc.wantErr)
				}
				return
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Unexpected object (-want, +got) = %v", diff)
			}
		})
	}
}

func TestURLString(t *testing.T) {
	testCases := map[string]struct {
		t    *URL
		want string
	}{
		"nil": {},
		"empty": {
			t:    &URL{},
			want: "",
		},
		"relative": {
			t:    &URL{Path: "/path/to/something"},
			want: "/path/to/something",
		},
		"nopath": {
			t:    HTTPS("foo"),
			want: "https://foo",
		},
		"absolute": {
			t: &URL{
				Scheme: "http",
				Host:   "path",
				Path:   "/to/something",
			},
			want: "http://path/to/something",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			got := tc.t

			if diff := cmp.Diff(tc.want, got.String()); diff != "" {
				t.Errorf("unexpected string (-want, +got) = %v", diff)
			}
			if diff := cmp.Diff(tc.want, got.URL().String()); diff != "" {
				t.Errorf("unexpected URL (-want, +got) = %v", diff)
			}
		})
	}
}

// These are lifted from the net/url url_test.go
var resolveReferenceTests = []struct {
	base, rel, expected string
}{
	// Absolute URL references
	{"http://foo.com?a=b", "https://bar.com/", "https://bar.com/"},
	{"http://foo.com/", "https://bar.com/?a=b", "https://bar.com/?a=b"},
	{"http://foo.com/", "https://bar.com/?", "https://bar.com/?"},
	{"http://foo.com/bar", "mailto:foo@example.com", "mailto:foo@example.com"},

	// Path-absolute references
	{"http://foo.com/bar", "/baz", "http://foo.com/baz"},
	{"http://foo.com/bar?a=b#f", "/baz", "http://foo.com/baz"},
	{"http://foo.com/bar?a=b", "/baz?", "http://foo.com/baz?"},
	{"http://foo.com/bar?a=b", "/baz?c=d", "http://foo.com/baz?c=d"},

	// Multiple slashes
	{"http://foo.com/bar", "http://foo.com//baz", "http://foo.com//baz"},
	{"http://foo.com/bar", "http://foo.com///baz/quux", "http://foo.com///baz/quux"},

	// Scheme-relative
	{"https://foo.com/bar?a=b", "//bar.com/quux", "https://bar.com/quux"},

	// Path-relative references:

	// ... current directory
	{"http://foo.com", ".", "http://foo.com/"},
	{"http://foo.com/bar", ".", "http://foo.com/"},
	{"http://foo.com/bar/", ".", "http://foo.com/bar/"},

	// ... going down
	{"http://foo.com", "bar", "http://foo.com/bar"},
	{"http://foo.com/", "bar", "http://foo.com/bar"},
	{"http://foo.com/bar/baz", "quux", "http://foo.com/bar/quux"},

	// ... going up
	{"http://foo.com/bar/baz", "../quux", "http://foo.com/quux"},
	{"http://foo.com/bar/baz", "../../../../../quux", "http://foo.com/quux"},
	{"http://foo.com/bar", "..", "http://foo.com/"},
	{"http://foo.com/bar/baz", "./..", "http://foo.com/"},
	// ".." in the middle (issue 3560)
	{"http://foo.com/bar/baz", "quux/dotdot/../tail", "http://foo.com/bar/quux/tail"},
	{"http://foo.com/bar/baz", "quux/./dotdot/../tail", "http://foo.com/bar/quux/tail"},
	{"http://foo.com/bar/baz", "quux/./dotdot/.././tail", "http://foo.com/bar/quux/tail"},
	{"http://foo.com/bar/baz", "quux/./dotdot/./../tail", "http://foo.com/bar/quux/tail"},
	{"http://foo.com/bar/baz", "quux/./dotdot/dotdot/././../../tail", "http://foo.com/bar/quux/tail"},
	{"http://foo.com/bar/baz", "quux/./dotdot/dotdot/./.././../tail", "http://foo.com/bar/quux/tail"},
	{"http://foo.com/bar/baz", "quux/./dotdot/dotdot/dotdot/./../../.././././tail", "http://foo.com/bar/quux/tail"},
	{"http://foo.com/bar/baz", "quux/./dotdot/../dotdot/../dot/./tail/..", "http://foo.com/bar/quux/dot/"},

	// Remove any dot-segments prior to forming the target URI.
	// http://tools.ietf.org/html/rfc3986#section-5.2.4
	{"http://foo.com/dot/./dotdot/../foo/bar", "../baz", "http://foo.com/dot/baz"},

	// Triple dot isn't special
	{"http://foo.com/bar", "...", "http://foo.com/..."},

	// Fragment
	{"http://foo.com/bar", ".#frag", "http://foo.com/#frag"},
	{"http://example.org/", "#!$&%27()*+,;=", "http://example.org/#!$&%27()*+,;="},

	// Paths with escaping (issue 16947).
	{"http://foo.com/foo%2fbar/", "../baz", "http://foo.com/baz"},
	{"http://foo.com/1/2%2f/3%2f4/5", "../../a/b/c", "http://foo.com/1/a/b/c"},
	{"http://foo.com/1/2/3", "./a%2f../../b/..%2fc", "http://foo.com/1/2/b/..%2fc"},
	{"http://foo.com/1/2%2f/3%2f4/5", "./a%2f../b/../c", "http://foo.com/1/2%2f/3%2f4/a%2f../c"},
	{"http://foo.com/foo%20bar/", "../baz", "http://foo.com/baz"},
	{"http://foo.com/foo", "../bar%2fbaz", "http://foo.com/bar%2fbaz"},
	{"http://foo.com/foo%2dbar/", "./baz-quux", "http://foo.com/foo%2dbar/baz-quux"},

	// RFC 3986: Normal Examples
	// http://tools.ietf.org/html/rfc3986#section-5.4.1
	{"http://a/b/c/d;p?q", "g:h", "g:h"},
	{"http://a/b/c/d;p?q", "g", "http://a/b/c/g"},
	{"http://a/b/c/d;p?q", "./g", "http://a/b/c/g"},
	{"http://a/b/c/d;p?q", "g/", "http://a/b/c/g/"},
	{"http://a/b/c/d;p?q", "/g", "http://a/g"},
	{"http://a/b/c/d;p?q", "//g", "http://g"},
	{"http://a/b/c/d;p?q", "?y", "http://a/b/c/d;p?y"},
	{"http://a/b/c/d;p?q", "g?y", "http://a/b/c/g?y"},
	{"http://a/b/c/d;p?q", "#s", "http://a/b/c/d;p?q#s"},
	{"http://a/b/c/d;p?q", "g#s", "http://a/b/c/g#s"},
	{"http://a/b/c/d;p?q", "g?y#s", "http://a/b/c/g?y#s"},
	{"http://a/b/c/d;p?q", ";x", "http://a/b/c/;x"},
	{"http://a/b/c/d;p?q", "g;x", "http://a/b/c/g;x"},
	{"http://a/b/c/d;p?q", "g;x?y#s", "http://a/b/c/g;x?y#s"},
	{"http://a/b/c/d;p?q", "", "http://a/b/c/d;p?q"},
	{"http://a/b/c/d;p?q", ".", "http://a/b/c/"},
	{"http://a/b/c/d;p?q", "./", "http://a/b/c/"},
	{"http://a/b/c/d;p?q", "..", "http://a/b/"},
	{"http://a/b/c/d;p?q", "../", "http://a/b/"},
	{"http://a/b/c/d;p?q", "../g", "http://a/b/g"},
	{"http://a/b/c/d;p?q", "../..", "http://a/"},
	{"http://a/b/c/d;p?q", "../../", "http://a/"},
	{"http://a/b/c/d;p?q", "../../g", "http://a/g"},

	// RFC 3986: Abnormal Examples
	// http://tools.ietf.org/html/rfc3986#section-5.4.2
	{"http://a/b/c/d;p?q", "../../../g", "http://a/g"},
	{"http://a/b/c/d;p?q", "../../../../g", "http://a/g"},
	{"http://a/b/c/d;p?q", "/./g", "http://a/g"},
	{"http://a/b/c/d;p?q", "/../g", "http://a/g"},
	{"http://a/b/c/d;p?q", "g.", "http://a/b/c/g."},
	{"http://a/b/c/d;p?q", ".g", "http://a/b/c/.g"},
	{"http://a/b/c/d;p?q", "g..", "http://a/b/c/g.."},
	{"http://a/b/c/d;p?q", "..g", "http://a/b/c/..g"},
	{"http://a/b/c/d;p?q", "./../g", "http://a/b/g"},
	{"http://a/b/c/d;p?q", "./g/.", "http://a/b/c/g/"},
	{"http://a/b/c/d;p?q", "g/./h", "http://a/b/c/g/h"},
	{"http://a/b/c/d;p?q", "g/../h", "http://a/b/c/h"},
	{"http://a/b/c/d;p?q", "g;x=1/./y", "http://a/b/c/g;x=1/y"},
	{"http://a/b/c/d;p?q", "g;x=1/../y", "http://a/b/c/y"},
	{"http://a/b/c/d;p?q", "g?y/./x", "http://a/b/c/g?y/./x"},
	{"http://a/b/c/d;p?q", "g?y/../x", "http://a/b/c/g?y/../x"},
	{"http://a/b/c/d;p?q", "g#s/./x", "http://a/b/c/g#s/./x"},
	{"http://a/b/c/d;p?q", "g#s/../x", "http://a/b/c/g#s/../x"},

	// Extras.
	{"https://a/b/c/d;p?q", "//g?q", "https://g?q"},
	{"https://a/b/c/d;p?q", "//g#s", "https://g#s"},
	{"https://a/b/c/d;p?q", "//g/d/e/f?y#s", "https://g/d/e/f?y#s"},
	{"https://a/b/c/d;p#s", "?y", "https://a/b/c/d;p?y"},
	{"https://a/b/c/d;p?q#s", "?y", "https://a/b/c/d;p?y"},
}

func TestResolveReference(t *testing.T) {
	mustParse := func(url string) *URL {
		u, err := ParseURL(url)
		if err != nil {
			t.Fatalf("ParseURL(%q) got err %v", url, err)
		}
		return u
	}

	for _, tc := range resolveReferenceTests {
		apisURL := mustParse(tc.base)
		apisRel := mustParse(tc.rel)
		if got := apisURL.ResolveReference(apisRel).String(); got != tc.expected {
			t.Errorf("URL(%q).ResolveReference(%q)\ngot  %q\nwant %q", tc.base, tc.rel, got, tc.expected)
		}
	}
}

func TestSemanticEquality(t *testing.T) {
	u1, err := ParseURL("https://user:password@example.com")
	if err != nil {
		t.Fatalf("ParseURL() got err %v", err)
	}

	u2, err := ParseURL("https://user:password@example.com")
	if err != nil {
		t.Fatalf("ParseURL() got err %v", err)
	}

	u3, err := ParseURL("https://another-user:password@example.com")
	if err != nil {
		t.Fatalf("ParseURL() got err %v", err)
	}

	if !equality.Semantic.DeepEqual(u1, u2) {
		t.Errorf("expected urls to be equivalent")
	}

	if equality.Semantic.DeepEqual(u1, u3) {
		t.Errorf("expected urls to be different")
	}
}
