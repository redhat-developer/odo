// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package integration

import (
	"context"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
)

//
// repository sub-tests
//

func testRepos(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		t.Run("Find", testRepoFind(client))
	}
}

func testRepoFind(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		result, _, err := client.Repositories.Find(context.Background(), "diaspora/diaspora")
		if err != nil {
			t.Error(err)
			return
		}
		t.Run("Repository", testRepo(result))
		t.Run("Permissions", testPerm(result.Perm))
	}
}

//
// struct sub-tests
//

func testRepo(repository *scm.Repository) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		if got, want := repository.Name, "diaspora"; got != want {
			t.Errorf("Want repository Name %q, got %q", want, got)
		}
		if got, want := repository.Namespace, "diaspora"; got != want {
			t.Errorf("Want repository Namespace %q, got %q", want, got)
		}
		if got, want := repository.Branch, "master"; got != want {
			t.Errorf("Want repository Branch %q, got %q", want, got)
		}
		if got, want := repository.Private, false; got != want {
			t.Errorf("Want repository Private %v, got %v", want, got)
		}
	}
}

func testPerm(perms *scm.Perm) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		if got, want := perms.Pull, true; got != want {
			t.Errorf("Want permission Pull %v, got %v", want, got)
		}
		if got, want := perms.Push, false; got != want {
			t.Errorf("Want permission Push %v, got %v", want, got)
		}
		if got, want := perms.Admin, false; got != want {
			t.Errorf("Want permission Admin %v, got %v", want, got)
		}
	}
}
