// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gitea

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/h2non/gock"
	"github.com/jenkins-x/go-scm/scm"
)

//
// commit sub-tests
//

func TestCommitFind(t *testing.T) {
	gock.New("https://try.gitea.io").
		Get("/api/v1/repos/gitea/gitea/git/commits/c43399cad8766ee521b873a32c1652407c5a4630").
		Reply(200).
		Type("application/json").
		File("testdata/commits.json")

	client, _ := New("https://try.gitea.io")
	got, _, err := client.Git.FindCommit(
		context.Background(),
		"gitea/gitea",
		"c43399cad8766ee521b873a32c1652407c5a4630",
	)
	if err != nil {
		t.Error(err)
	}

	want := new(scm.Commit)
	raw, _ := ioutil.ReadFile("testdata/commits.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestCommitList(t *testing.T) {
	client, _ := New("https://try.gitea.io")
	_, _, err := client.Git.ListCommits(context.Background(), "go-gitea/gitea", scm.CommitListOptions{})
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestChangeList(t *testing.T) {
	client, _ := New("https://try.gitea.io")
	_, _, err := client.Git.ListChanges(context.Background(), "go-gitea/gitea", "f05f642b892d59a0a9ef6a31f6c905a24b5db13a", scm.ListOptions{})
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

//
// branch sub-tests
//

func TestBranchFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://try.gitea.io").
		Get("/api/v1/repos/go-gitea/gitea/branches/master").
		Reply(200).
		Type("application/json").
		File("testdata/branch.json")

	client, _ := New("https://try.gitea.io")
	got, _, err := client.Git.FindBranch(context.Background(), "go-gitea/gitea", "master")
	if err != nil {
		t.Error(err)
	}

	want := new(scm.Reference)
	raw, _ := ioutil.ReadFile("testdata/branch.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestBranchList(t *testing.T) {
	defer gock.Off()

	gock.New("https://try.gitea.io").
		Get("/api/v1/repos/go-gitea/gitea/branches").
		Reply(200).
		Type("application/json").
		File("testdata/branches.json")

	client, _ := New("https://try.gitea.io")
	got, _, err := client.Git.ListBranches(context.Background(), "go-gitea/gitea", scm.ListOptions{})
	if err != nil {
		t.Error(err)
	}

	want := []*scm.Reference{}
	raw, _ := ioutil.ReadFile("testdata/branches.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

//
// tag sub-tests
//

func TestTagFind(t *testing.T) {
	client, _ := New("https://try.gitea.io")
	_, _, err := client.Git.FindTag(context.Background(), "go-gitea/gitea", "v1.0.0")
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestTagList(t *testing.T) {
	client, _ := New("https://try.gitea.io")
	_, _, err := client.Git.ListTags(context.Background(), "go-gitea/gitea", scm.ListOptions{})
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}
