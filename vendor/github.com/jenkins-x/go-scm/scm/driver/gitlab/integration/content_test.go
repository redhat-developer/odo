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
		result, _, err := client.Contents.Find(context.Background(), "gitlab-org/testme", "VERSION", "5937ac0a7beb003549fc5fd26fc247adbce4a52e")
		if err != nil {
			t.Error(err)
			return
		}
		if got, want := result.Path, "VERSION"; got != want {
			t.Errorf("Got file path %q, want %q", got, want)
		}
		if got, want := string(result.Data), "6.7.0.pre\n"; got != want {
			t.Errorf("Got file data %q, want %q", got, want)
		}
	}
}

func testContentFindBranch(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		result, _, err := client.Contents.Find(context.Background(), "gitlab-org/testme", "VERSION", "feature")
		if err != nil {
			t.Error(err)
			return
		}
		if got, want := result.Path, "VERSION"; got != want {
			t.Errorf("Got file path %q, want %q", got, want)
		}
		if got, want := string(result.Data), "6.7.0.pre\n"; got != want {
			t.Errorf("Got file data %q, want %q", got, want)
		}
	}
}
