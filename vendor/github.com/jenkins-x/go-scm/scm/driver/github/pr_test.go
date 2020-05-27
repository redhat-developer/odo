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
		File("testdata/close.json").
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

func TestPullReopen(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Patch("/repos/octocat/hello-world/pulls/1347").
		File("testdata/reopen.json").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders)

	client := NewDefault()
	res, err := client.PullRequests.Reopen(context.Background(), "octocat/hello-world", 1347)
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

func TestPullUpdate(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Patch("/repos/octocat/hello-world/pulls/1").
		File("testdata/pr_update.json").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/pr_create.json")

	input := &scm.PullRequestInput{
		Title: "A new title",
		Body:  "A new description",
	}

	client := NewDefault()

	got, res, err := client.PullRequests.Update(context.Background(), "octocat/hello-world", 1, input)
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

func TestPullService_RequestReview(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Post("/repos/octocat/hello-world/pulls/1/requested_reviewers").
		File("testdata/request_reviewers.json").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/pr_create.json")

	logins := []string{"octocat", "hubot", "other_user", "octocat/justice-league"}
	client := NewDefault()

	res, err := client.PullRequests.RequestReview(context.Background(), "octocat/hello-world", 1, logins)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))

}

func TestPullService_UnrequestReview(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Delete("/repos/octocat/hello-world/pulls/1/requested_reviewers").
		File("testdata/delete_reviewers.json").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/delete_reviewers_pr.json")

	logins := []string{"octocat", "hubot", "other_user", "octocat/justice-league"}
	client := NewDefault()

	res, err := client.PullRequests.UnrequestReview(context.Background(), "octocat/hello-world", 1, logins)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))

}
