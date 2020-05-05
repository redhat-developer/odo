// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package integration

import (
	"context"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
)

func testContents(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		t.Run("Find", testContentFind(client))
		t.Run("Find/Branch", testContentFindBranch(client))
	}
}

func testContentFind(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		result, _, err := client.Contents.Find(context.Background(), "octocat/Hello-World", "README", "7fd1a60b01f91b314f59955a4e4d4e80d8edf11d")
		if err != nil {
			t.Error(err)
			return
		}
		if got, want := result.Path, "README"; got != want {
			t.Errorf("Got file path %q, want %q", got, want)
		}
		if got, want := string(result.Data), "Hello World!\n"; got != want {
			t.Errorf("Got file data %q, want %q", got, want)
		}
	}
}

func testContentFindBranch(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		result, _, err := client.Contents.Find(context.Background(), "octocat/Hello-World", "CONTRIBUTING.md", "test")
		if err != nil {
			t.Error(err)
			return
		}
		if got, want := result.Path, "CONTRIBUTING.md"; got != want {
			t.Errorf("Got file path %q, want %q", got, want)
		}
		if got, want := string(result.Data), "## Contributing\n"; got != want {
			t.Errorf("Got file data %q, want %q", got, want)
		}
	}
}
