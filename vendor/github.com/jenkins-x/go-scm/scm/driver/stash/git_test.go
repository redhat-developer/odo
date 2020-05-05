// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stash

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/jenkins-x/go-scm/scm"

	"github.com/google/go-cmp/cmp"
	"github.com/h2non/gock"
)

func TestGitFindCommit(t *testing.T) {
	defer gock.Off()

	gock.New("http://example.com:7990").
		Get("/rest/api/1.0/projects/PRJ/repos/my-repo/commits/131cb13f4aed12e725177bc4b7c28db67839bf9f").
		Reply(200).
		Type("application/json").
		File("testdata/commit.json")

	client, _ := New("http://example.com:7990")
	got, _, err := client.Git.FindCommit(context.Background(), "PRJ/my-repo", "131cb13f4aed12e725177bc4b7c28db67839bf9f")
	if err != nil {
		t.Error(err)
	}

	want := new(scm.Commit)
	raw, _ := ioutil.ReadFile("testdata/commit.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestGitFindBranch(t *testing.T) {
	defer gock.Off()

	gock.New("http://example.com:7990").
		Get("/rest/api/1.0/projects/PRJ/repos/my-repo/branches").
		MatchParam("filterText", "master").
		Reply(200).
		Type("application/json").
		File("testdata/branch.json")

	client, _ := New("http://example.com:7990")
	got, _, err := client.Git.FindBranch(context.Background(), "PRJ/my-repo", "master")
	if err != nil {
		t.Error(err)
	}

	want := new(scm.Reference)
	raw, _ := ioutil.ReadFile("testdata/branch.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestGitFindTag(t *testing.T) {
	defer gock.Off()

	gock.New("http://example.com:7990").
		Get("/rest/api/1.0/projects/PRJ/repos/my-repo/tags").
		MatchParam("filterText", "v1.0.0").
		Reply(200).
		Type("application/json").
		File("testdata/tag.json")

	client, _ := New("http://example.com:7990")
	got, _, err := client.Git.FindTag(context.Background(), "PRJ/my-repo", "v1.0.0")
	if err != nil {
		t.Error(err)
	}

	want := new(scm.Reference)
	raw, _ := ioutil.ReadFile("testdata/tag.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestGitListCommits(t *testing.T) {
	client, _ := New("http://example.com:7990")
	_, _, err := client.Git.ListCommits(context.Background(), "PRJ/my-repo", scm.CommitListOptions{Ref: "master", Page: 1, Size: 30})
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestGitListBranches(t *testing.T) {
	defer gock.Off()

	gock.New("http://example.com:7990").
		Get("/rest/api/1.0/projects/PRJ/repos/my-repo/branches").
		MatchParam("limit", "30").
		Reply(200).
		Type("application/json").
		File("testdata/branches.json")

	client, _ := New("http://example.com:7990")
	got, _, err := client.Git.ListBranches(context.Background(), "PRJ/my-repo", scm.ListOptions{Page: 1, Size: 30})
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
	//
	// t.Run("Page", testPage(res))
}

func TestGitListTags(t *testing.T) {
	defer gock.Off()

	gock.New("http://example.com:7990").
		Get("/rest/api/1.0/projects/PRJ/repos/my-repo/tags").
		MatchParam("limit", "30").
		Reply(200).
		Type("application/json").
		File("testdata/tags.json")

	client, _ := New("http://example.com:7990")
	got, _, err := client.Git.ListTags(context.Background(), "PRJ/my-repo", scm.ListOptions{Page: 1, Size: 30})
	if err != nil {
		t.Error(err)
	}

	want := []*scm.Reference{}
	raw, _ := ioutil.ReadFile("testdata/tags.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	// t.Run("Page", testPage(res))
}

func TestGitListChanges(t *testing.T) {
	defer gock.Off()

	gock.New("http://example.com:7990").
		Get("/rest/api/1.0/projects/PRJ/repos/my-repo/commits/131cb13f4aed12e725177bc4b7c28db67839bf9f/changes").
		MatchParam("limit", "30").
		Reply(200).
		Type("application/json").
		File("testdata/changes.json")

	client, _ := New("http://example.com:7990")
	got, _, err := client.Git.ListChanges(context.Background(), "PRJ/my-repo", "131cb13f4aed12e725177bc4b7c28db67839bf9f", scm.ListOptions{Page: 1, Size: 30})
	if err != nil {
		t.Error(err)
	}

	want := []*scm.Change{}
	raw, _ := ioutil.ReadFile("testdata/changes.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}
