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

package helpers

import (
	"regexp"
	"testing"
)

var matcher = regexp.MustCompile("abcd-[a-z]{8}")

func TestAppendRandomString(t *testing.T) {
	const s = "abcd"
	w := AppendRandomString(s)
	o := AppendRandomString(s)
	if !matcher.MatchString(w) || !matcher.MatchString(o) || o == w {
		t.Fatalf("Generated string(s) are incorrect: %q, %q", w, o)
	}
}

func TestMakeK8sNamePrefix(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"abcd123", "abcd123"},
		{"AbCdef", "ab-cdef"},
		{"ABCD", "a-b-c-d"},
		{"aBc*ef&d", "a-bc-ef-d"},
	}
	for _, v := range testCases {
		actual := MakeK8sNamePrefix(v.input)
		if v.expected != actual {
			t.Fatalf("Expect %q but actual is %q", v.expected, actual)
		}
	}
}

func TestGetBaseFuncName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"test/e2e.TestMain", "TestMain"},
		{"e2e.TestMain", "TestMain"},
		{"test/TestMain", "TestMain"},
		{"TestMain", "TestMain"},
	}
	for _, v := range testCases {
		actual := GetBaseFuncName(v.input)
		if v.expected != actual {
			t.Fatalf("Expect %q but actual is %q", v.expected, actual)
		}
	}
}
