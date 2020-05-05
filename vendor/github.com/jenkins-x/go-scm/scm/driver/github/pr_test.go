// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/jenkins-x/go-scm/scm"

	"github.com/google/go-cmp/cmp"
	"github.com/h2non/gock"
)

func TestPullFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/repos/octocat/hello-world/pulls/1347").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/pr.json")

	client := NewDefault()
	got, res, err := client.PullRequests.Find(context.Background(), "octocat/hello-world", 1347)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.PullRequest)
	raw, _ := ioutil.ReadFile("testdata/pr.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestPullList(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/repos/octocat/hello-world/pulls").
		MatchParam("page", "1").
		MatchParam("per_page", "30").
		MatchParam("state", "all").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/pulls.json")

	client := NewDefault()
	got, res, err := client.PullRequests.List(context.Background(), "octocat/hello-world", scm.PullRequestListOptions{Page: 1, Size: 30, Open: true, Closed: true})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.PullRequest{}
	raw, _ := ioutil.ReadFile("testdata/pulls.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
	t.Run("Page", testPage(res))
}

func TestPullListChanges(t *testing.T) {
	gock.New("https://api.github.com").
		Get("/repos/octocat/hello-world/pulls/1347/files").
		MatchParam("page", "1").
		MatchParam("per_page", "30").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/pr_files.json")

	client := NewDefault()
	got, res, err := client.PullRequests.ListChanges(context.Background(), "octocat/hello-world", 1347, scm.ListOptions{Page: 1, Size: 30})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Change{}
	raw, _ := ioutil.ReadFile("testdata/pr_files.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
	t.Run("Page", testPage(res))
}

func TestPullMerge(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Put("/repos/octocat/hello-world/pulls/1347/merge").
		File("testdata/pr_merge.json").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders)

	client := NewDefault()
	mergeOptions := &scm.PullRequestMergeOptions{
		MergeMethod: "rebase",
	}
	res, err := client.PullRequests.Merge(context.Background(), "octocat/hello-world", 1347, mergeOptions)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestPullClose(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Patch("/repos/octocat/hello-world/pulls/1347").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders)

	client := NewDefault()
	res, err := client.PullRequests.Close(context.Background(), "octocat/hello-world", 1347)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestPullCreate(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Post("/repos/octocat/hello-world/pulls").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/pr_create.json")

	input := &scm.PullRequestInput{
		Title: "Amazing new feature",
		Body:  "Please pull these awesome changes in!",
		Head:  "octocat:new-feature",
		Base:  "master",
	}

	client := NewDefault()

	got, res, err := client.PullRequests.Create(context.Background(), "octocat/hello-world", input)
	if err != nil {
		t.Fatal(err)
	}

	want := new(scm.PullRequest)
	raw, err := ioutil.ReadFile("testdata/pr_create.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}
