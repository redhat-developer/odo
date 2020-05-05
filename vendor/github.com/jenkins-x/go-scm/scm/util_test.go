// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		value, owner, name string
	}{
		{"octocat/hello-world", "octocat", "hello-world"},
		{"octocat/hello/world", "octocat", "hello/world"},
		{"hello-world", "", "hello-world"},
		{value: ""}, // empty value returns nothing
	}
	for _, test := range tests {
		owner, name := Split(test.value)
		if got, want := owner, test.owner; got != want {
			t.Errorf("Got repository owner %s, want %s", got, want)
		}
		if got, want := name, test.name; got != want {
			t.Errorf("Got repository name %s, want %s", got, want)
		}
	}
}

func TestJoin(t *testing.T) {
	got, want := Join("octocat", "hello-world"), "octocat/hello-world"
	if got != want {
		t.Errorf("Got repository name %s, want %s", got, want)
	}
}

func TestUrlJoin(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "http://foo.bar/whatnot/thingy", UrlJoin("http://foo.bar", "whatnot", "thingy"))
	assert.Equal(t, "http://foo.bar/whatnot/thingy/", UrlJoin("http://foo.bar/", "/whatnot/", "/thingy/"))
}

func TestTrimRef(t *testing.T) {
	tests := []struct {
		before, after string
	}{
		{
			before: "refs/tags/v1.0.0",
			after:  "v1.0.0",
		},
		{
			before: "refs/heads/master",
			after:  "master",
		},
		{
			before: "refs/heads/feature/x",
			after:  "feature/x",
		},
		{
			before: "master",
			after:  "master",
		},
	}
	for _, test := range tests {
		if got, want := TrimRef(test.before), test.after; got != want {
			t.Errorf("Got reference %s, want %s", got, want)
		}
	}
}

func TestExpandRef(t *testing.T) {
	tests := []struct {
		name, prefix, after string
	}{
		// tag references
		{
			after:  "refs/tags/v1.0.0",
			name:   "v1.0.0",
			prefix: "refs/tags",
		},
		{
			after:  "refs/tags/v1.0.0",
			name:   "v1.0.0",
			prefix: "refs/tags/",
		},
		// branch references
		{
			after:  "refs/heads/master",
			name:   "master",
			prefix: "refs/heads",
		},
		{
			after:  "refs/heads/master",
			name:   "master",
			prefix: "refs/heads/",
		},
		// is already a ref
		{
			after:  "refs/tags/v1.0.0",
			name:   "refs/tags/v1.0.0",
			prefix: "refs/heads/",
		},
	}
	for _, test := range tests {
		if got, want := ExpandRef(test.name, test.prefix), test.after; got != want {
			t.Errorf("Got reference %s, want %s", got, want)
		}
	}
}

func TestIsRef(t *testing.T) {
	tests := []struct {
		name string
		tag  bool
	}{
		// tag references
		{
			name: "refs/tags/v1.0.0",
			tag:  true,
		},
		{
			name: "refs/heads/master",
			tag:  false,
		},
	}
	for _, test := range tests {
		if got, want := IsTag(test.name), test.tag; got != want {
			t.Errorf("Got IsTag %v, want %v", got, want)
		}
	}
}
