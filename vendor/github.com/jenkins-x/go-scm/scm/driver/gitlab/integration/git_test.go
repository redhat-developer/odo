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
		result, _, err := client.Git.FindBranch(context.Background(), "gitlab-org/testme", "feature")
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
		result, _, err := client.Git.ListBranches(context.Background(), "gitlab-org/testme", opts)
		if err != nil {
			t.Error(err)
			return
		}
		if len(result) == 0 {
			t.Errorf("Want a non-empty branch list")
		}
		for _, branch := range result {
			if branch.Name == "feature" {
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
		t.Parallel()
		result, _, err := client.Git.FindTag(context.Background(), "gitlab-org/testme", "v1.1.0")
		if err != nil {
			t.Error(err)
			return
		}
		t.Run("Tag", testTag(result))
	}
}

func testTagList(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		opts := scm.ListOptions{}
		result, _, err := client.Git.ListTags(context.Background(), "gitlab-org/testme", opts)
		if err != nil {
			t.Error(err)
			return
		}
		if len(result) == 0 {
			t.Errorf("Want a non-empty tag list")
		}
		for _, tag := range result {
			if tag.Name == "v1.1.0" {
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
		result, _, err := client.Git.FindCommit(context.Background(), "gitlab-org/testme", "0b4bc9a49b562e85de7cc9e834518ea6828729b9")
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
			Ref: "feature",
		}
		result, _, err := client.Git.ListCommits(context.Background(), "gitlab-org/testme", opts)
		if err != nil {
			t.Error(err)
			return
		}
		if len(result) == 0 {
			t.Errorf("Want a non-empty commit list")
		}
		for _, commit := range result {
			if commit.Sha == "0b4bc9a49b562e85de7cc9e834518ea6828729b9" {
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
		if got, want := branch.Name, "feature"; got != want {
			t.Errorf("Want branch Name %q, got %q", want, got)
		}
		if got, want := branch.Sha, "0b4bc9a49b562e85de7cc9e834518ea6828729b9"; got != want {
			t.Errorf("Want branch Avatar %q, got %q", want, got)
		}
	}
}

func testTag(tag *scm.Reference) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := tag.Name, "v1.1.0"; got != want {
			t.Errorf("Want tag Name %q, got %q", want, got)
		}
		if got, want := tag.Sha, "5937ac0a7beb003549fc5fd26fc247adbce4a52e"; got != want {
			t.Errorf("Want tag Avatar %q, got %q", want, got)
		}
	}
}

func testCommit(commit *scm.Commit) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := commit.Message, "Feature added\n\nSigned-off-by: Dmitriy Zaporozhets <dmitriy.zaporozhets@gmail.com>\n"; got != want {
			t.Errorf("Want commit Message %q, got %q", want, got)
		}
		if got, want := commit.Sha, "0b4bc9a49b562e85de7cc9e834518ea6828729b9"; got != want {
			t.Errorf("Want commit Sha %q, got %q", want, got)
		}
		if got, want := commit.Author.Name, "Dmitriy Zaporozhets"; got != want {
			t.Errorf("Want commit author Name %q, got %q", want, got)
		}
		if got, want := commit.Author.Email, "dmitriy.zaporozhets@gmail.com"; got != want {
			t.Errorf("Want commit author Email %q, got %q", want, got)
		}
		if got, want := commit.Author.Date.Unix(), int64(1393489561); got != want {
			t.Errorf("Want commit author Date %d, got %d", want, got)
		}
		if got, want := commit.Committer.Name, "Dmitriy Zaporozhets"; got != want {
			t.Errorf("Want commit author Name %q, got %q", want, got)
		}
		if got, want := commit.Committer.Email, "dmitriy.zaporozhets@gmail.com"; got != want {
			t.Errorf("Want commit author Email %q, got %q", want, got)
		}
		if got, want := commit.Committer.Date.Unix(), int64(1393489561); got != want {
			t.Errorf("Want commit author Date %d, got %d", want, got)
		}
	}
}
