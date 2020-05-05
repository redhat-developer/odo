// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stash

import (
	"testing"

	"github.com/jenkins-x/go-scm/scm"
)

func Test_encodeListOptions(t *testing.T) {
	tests := []struct {
		page int
		size int
		text string
	}{
		{page: 0, size: 30, text: "limit=30"},
		{page: 1, size: 30, text: "limit=30"},
		{page: 5, size: 30, text: "limit=30&start=120"},
		{page: 2, size: 5, text: "limit=5&start=5"},
	}
	for _, test := range tests {
		opts := scm.ListOptions{
			Page: test.page,
			Size: test.size,
		}
		if got, want := encodeListOptions(opts), test.text; got != want {
			t.Errorf("Want encoded list options %q, got %q", want, got)
		}
	}
}

func Test_encodePullRequestListOptions(t *testing.T) {
	t.Parallel()
	opts := scm.PullRequestListOptions{
		Page:   10,
		Size:   30,
		Open:   true,
		Closed: true,
	}
	want := "limit=30&start=270&state=all"
	got := encodePullRequestListOptions(opts)
	if got != want {
		t.Errorf("Want encoded pr list options %q, got %q", want, got)
	}
}
