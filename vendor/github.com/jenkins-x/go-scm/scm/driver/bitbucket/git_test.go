// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitbucket

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

	gock.New("https://api.bitbucket.org").
		Get("/2.0/repositories/atlassian/stash-example-plugin/commit/a6e5e7d797edf751cbd839d6bd4aef86c941eec9").
		Reply(200).
		Type("application/json").
		File("testdata/commit.json")

	client, _ := New("https://api.bitbucket.org")
	got, _, err := client.Git.FindCommit(context.Background(), "atlassian/stash-example-plugin", "a6e5e7d797edf751cbd839d6bd4aef86c941eec9")
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

	gock.New("https://api.bitbucket.org").
		Get("/2.0/repositories/atlassian/stash-example-plugin/refs/branches/master").
		Reply(200).
		Type("application/json").
		File("testdata/branch.json")

	client, _ := New("https://api.bitbucket.org")
	got, _, err := client.Git.FindBranch(context.Background(), "atlassian/stash-example-plugin", "master")
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

	gock.New("https://api.bitbucket.org").
		Get("/2.0/repositories/atlassian/atlaskit/refs/tags/@atlaskit/activity@1.0.3").
		Reply(200).
		Type("application/json").
		File("testdata/tag.json")

	client, _ := New("https://api.bitbucket.org")
	got, _, err := client.Git.FindTag(context.Background(), "atlassian/atlaskit", "@atlaskit/activity@1.0.3")
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
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Get("/2.0/repositories/atlassian/stash-example-plugin/commits/master").
		MatchParam("page", "1").
		MatchParam("pagelen", "30").
		Reply(200).
		Type("application/json").
		File("testdata/commits.json")

	client, _ := New("https://api.bitbucket.org")
	got, res, err := client.Git.ListCommits(context.Background(), "atlassian/stash-example-plugin", scm.CommitListOptions{Ref: "master", Page: 1, Size: 30})
	if err != nil {
		t.Error(err)
	}

	want := []*scm.Commit{}
	raw, _ := ioutil.ReadFile("testdata/commits.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Page", testPage(res))
}

func TestGitListBranches(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Get("/2.0/repositories/atlassian/stash-example-plugin/refs/branches").
		MatchParam("page", "1").
		MatchParam("pagelen", "30").
		Reply(200).
		Type("application/json").
		File("testdata/branches.json")

	client, _ := New("https://api.bitbucket.org")
	got, res, err := client.Git.ListBranches(context.Background(), "atlassian/stash-example-plugin", scm.ListOptions{Page: 1, Size: 30})
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

	t.Run("Page", testPage(res))
}

func TestGitListTags(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Get("/2.0/repositories/atlassian/atlaskit/refs/tags").
		MatchParam("page", "1").
		MatchParam("pagelen", "30").
		Reply(200).
		Type("application/json").
		File("testdata/tags.json")

	client, _ := New("https://api.bitbucket.org")
	got, res, err := client.Git.ListTags(context.Background(), "atlassian/atlaskit", scm.ListOptions{Page: 1, Size: 30})
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

	t.Run("Page", testPage(res))
}

func TestGitListChanges(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Get("/2.0/repositories/atlassian/atlaskit/diffstat/425863f9dbe56d70c8dcdbf2e4e0805e85591fcc").
		MatchParam("page", "1").
		MatchParam("pagelen", "30").
		Reply(200).
		Type("application/json").
		File("testdata/diffstat.json")

	client, _ := New("https://api.bitbucket.org")
	got, _, err := client.Git.ListChanges(context.Background(), "atlassian/atlaskit", "425863f9dbe56d70c8dcdbf2e4e0805e85591fcc", scm.ListOptions{Page: 1, Size: 30})
	if err != nil {
		t.Error(err)
	}

	want := []*scm.Change{}
	raw, _ := ioutil.ReadFile("testdata/diffstat.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}
