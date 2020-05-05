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
// git sub-tests
//

func testGit(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		t.Run("Branches", testBranches(client))
		t.Run("Commits", testCommits(client))
		t.Run("Tags", testTags(client))
	}
}

//
// branch sub-tests
//

func testBranches(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		t.Run("Find", testBranchFind(client))
		t.Run("List", testBranchList(client))
	}
}

func testBranchFind(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		result, _, err := client.Git.FindBranch(context.Background(), "octocat/Hello-World", "master")
		if err != nil {
			t.Error(err)
			return
		}
		t.Run("Branch", testBranch(result))
	}
}

func testBranchList(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		opts := scm.ListOptions{}
		result, _, err := client.Git.ListBranches(context.Background(), "octocat/Hello-World", opts)
		if err != nil {
			t.Error(err)
			return
		}
		if len(result) == 0 {
			t.Errorf("Want a non-empty branch list")
		}
		for _, branch := range result {
			if branch.Name == "master" {
				t.Run("Branch", testBranch(branch))
			}
		}
	}
}

//
// branch sub-tests
//

func testTags(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		t.Run("Find", testTagFind(client))
		t.Run("List", testTagList(client))
	}
}

func testTagFind(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Skipf("Not Supported")
	}
}

func testTagList(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		opts := scm.ListOptions{}
		result, _, err := client.Git.ListTags(context.Background(), "octocat/linguist", opts)
		if err != nil {
			t.Error(err)
			return
		}
		if len(result) == 0 {
			t.Errorf("Want a non-empty tag list")
		}
		for _, tag := range result {
			if tag.Name == "v4.8.8" {
				t.Run("Tag", testTag(tag))
			}
		}
	}
}

//
// commit sub-tests
//

func testCommits(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		t.Run("Find", testCommitFind(client))
		t.Run("List", testCommitList(client))
	}
}

func testCommitFind(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		result, _, err := client.Git.FindCommit(context.Background(), "octocat/Hello-World", "7fd1a60b01f91b314f59955a4e4d4e80d8edf11d")
		if err != nil {
			t.Error(err)
			return
		}
		t.Run("Commit", testCommit(result))
	}
}

func testCommitList(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		opts := scm.CommitListOptions{
			Ref: "master",
		}
		result, _, err := client.Git.ListCommits(context.Background(), "octocat/Hello-World", opts)
		if err != nil {
			t.Error(err)
			return
		}
		if len(result) == 0 {
			t.Errorf("Want a non-empty commit list")
		}
		for _, commit := range result {
			if commit.Sha == "7fd1a60b01f91b314f59955a4e4d4e80d8edf11d" {
				t.Run("Commit", testCommit(commit))
			}
		}
	}
}

//
// struct sub-tests
//

func testBranch(branch *scm.Reference) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := branch.Name, "master"; got != want {
			t.Errorf("Want branch Name %q, got %q", want, got)
		}
		if got, want := branch.Sha, "7fd1a60b01f91b314f59955a4e4d4e80d8edf11d"; got != want {
			t.Errorf("Want branch Avatar %q, got %q", want, got)
		}
	}
}

func testTag(tag *scm.Reference) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := tag.Name, "v4.8.8"; got != want {
			t.Errorf("Want tag Name %q, got %q", want, got)
		}
		if got, want := tag.Sha, "3f4b8368e81430e3353cb5ad8b781cd044697347"; got != want {
			t.Errorf("Want tag Avatar %q, got %q", want, got)
		}
	}
}

func testCommit(commit *scm.Commit) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := commit.Message, "Merge pull request #6 from Spaceghost/patch-1\n\nNew line at end of file."; got != want {
			t.Errorf("Want commit Message %q, got %q", want, got)
		}
		if got, want := commit.Sha, "7fd1a60b01f91b314f59955a4e4d4e80d8edf11d"; got != want {
			t.Errorf("Want commit Sha %q, got %q", want, got)
		}
		if got, want := commit.Author.Name, "The Octocat"; got != want {
			t.Errorf("Want commit author Name %q, got %q", want, got)
		}
		if got, want := commit.Author.Email, "octocat@nowhere.com"; got != want {
			t.Errorf("Want commit author Email %q, got %q", want, got)
		}
		if got, want := commit.Author.Date.Unix(), int64(1331075210); got != want {
			t.Errorf("Want commit author Date %d, got %d", want, got)
		}
		if got, want := commit.Committer.Name, "The Octocat"; got != want {
			t.Errorf("Want commit author Name %q, got %q", want, got)
		}
		if got, want := commit.Committer.Email, "octocat@nowhere.com"; got != want {
			t.Errorf("Want commit author Email %q, got %q", want, got)
		}
		if got, want := commit.Committer.Date.Unix(), int64(1331075210); got != want {
			t.Errorf("Want commit author Date %d, got %d", want, got)
		}
	}
}
