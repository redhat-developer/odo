// Copyright (c) 2015 Dalton Hubble. All rights reserved.
// Copyrights licensed under the MIT License.

package oauth1

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEncodeParameterString(t *testing.T) {
	params := map[string]string{
		"key 1": "key 2",
		"key+3": "key+4",
	}
	want := "key%201=key%202&key%2B3=key%2B4"
	got := encodeParameterString(params)
	if got != want {
		t.Errorf("Want encoded string %s, got %s", want, got)
	}
}

func TestEncodeParameters(t *testing.T) {
	params := map[string]string{
		"key 1": "key 2",
		"key+3": "key+4",
	}
	want := map[string]string{
		"key%201": "key%202",
		"key%2B3": "key%2B4",
	}
	got := encodeParameters(params)
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestPercentEncode(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{" ", "%20"},
		{"%", "%25"},
		{"&", "%26"},
		{"-._", "-._"},
		{" /=+", "%20%2F%3D%2B"},
		{"Ladies + Gentlemen", "Ladies%20%2B%20Gentlemen"},
		{"An encoded string!", "An%20encoded%20string%21"},
		{"Dogs, Cats & Mice", "Dogs%2C%20Cats%20%26%20Mice"},
		{"☃", "%E2%98%83"},
	}
	for _, c := range cases {
		if output := percentEncode(c.input); output != c.expected {
			t.Errorf("expected %s, got %s", c.expected, output)
		}
	}
}
