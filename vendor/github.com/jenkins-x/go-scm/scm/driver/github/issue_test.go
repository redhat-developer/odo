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

func TestIssueFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/repos/octocat/hello-world/issues/1").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/issue.json")

	client := NewDefault()
	got, res, err := client.Issues.Find(context.Background(), "octocat/hello-world", 1)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Issue)
	raw, _ := ioutil.ReadFile("testdata/issue.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestIssueCommentFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/repos/octocat/hello-world/issues/comments/1").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/issue_comment.json")

	client := NewDefault()
	got, res, err := client.Issues.FindComment(context.Background(), "octocat/hello-world", 2, 1)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Comment)
	raw, _ := ioutil.ReadFile("testdata/issue_comment.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestIssueList(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/repos/octocat/hello-world/issues").
		MatchParam("page", "1").
		MatchParam("per_page", "30").
		MatchParam("state", "all").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/issues.json")

	client := NewDefault()
	got, res, err := client.Issues.List(context.Background(), "octocat/hello-world", scm.IssueListOptions{Page: 1, Size: 30, Open: true, Closed: true})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Issue{}
	raw, _ := ioutil.ReadFile("testdata/issues.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
	t.Run("Page", testPage(res))
}

func TestIssueSearch(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/search/issues").
		//MatchParam("page", "1").
		//MatchParam("per_page", "30").
		//MatchParam("state", "all").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/issue_search.json")

	client := NewDefault()
	searchOptions := scm.SearchOptions{
		Query:     "windows label:bug language:python state:open",
		Ascending: true,
	}
	got, res, err := client.Issues.Search(context.Background(), searchOptions)
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.SearchIssue{}
	raw, _ := ioutil.ReadFile("testdata/issue_search.json.golden")
	json.Unmarshal(raw, &want)

	data, _ := json.Marshal(got)
	t.Log(string(data))

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
	t.Run("Page", testPage(res))
}

func TestIssuePPRSearch(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/search/issues").
		//MatchParam("page", "1").
		//MatchParam("per_page", "30").
		//MatchParam("state", "all").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/issue_search_prs.json")

	client := NewDefault()
	searchOptions := scm.SearchOptions{
		Query:     "is:pr is:open author:jstrachan user:jenkins-x",
		Ascending: true,
	}
	got, res, err := client.Issues.Search(context.Background(), searchOptions)
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.SearchIssue{}
	raw, _ := ioutil.ReadFile("testdata/issue_search_prs.json.golden")
	json.Unmarshal(raw, &want)

	data, _ := json.Marshal(got)
	t.Log(string(data))

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
	t.Run("Page", testPage(res))
}

func TestIssueListComments(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/repos/octocat/hello-world/issues/1/comments").
		MatchParam("page", "1").
		MatchParam("per_page", "30").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/issue_comments.json")

	client := NewDefault()
	got, res, err := client.Issues.ListComments(context.Background(), "octocat/hello-world", 1, scm.ListOptions{Size: 30, Page: 1})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Comment{}
	raw, _ := ioutil.ReadFile("testdata/issue_comments.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
	t.Run("Page", testPage(res))
}

func TestIssueCreate(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Post("/repos/octocat/hello-world/issues").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/issue.json")

	input := scm.IssueInput{
		Title: "Found a bug",
		Body:  "I'm having a problem with this.",
	}

	client := NewDefault()
	got, res, err := client.Issues.Create(context.Background(), "octocat/hello-world", &input)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Issue)
	raw, _ := ioutil.ReadFile("testdata/issue.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestIssueCreateComment(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Post("/repos/octocat/hello-world/issues/1/comments").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/issue_comment.json")

	input := &scm.CommentInput{
		Body: "what?",
	}

	client := NewDefault()
	got, res, err := client.Issues.CreateComment(context.Background(), "octocat/hello-world", 1, input)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Comment)
	raw, _ := ioutil.ReadFile("testdata/issue_comment.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestIssueCommentDelete(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Delete("/repos/octocat/hello-world/issues/comments/1").
		Reply(204).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/issue.json")

	client := NewDefault()
	res, err := client.Issues.DeleteComment(context.Background(), "octocat/hello-world", 1, 1)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestIssueEditComment(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Patch("/repos/octocat/hello-world/issues/comments/1").
		File("testdata/edit_issue_comment.json").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/issue_comment.json")

	input := &scm.CommentInput{
		Body: "what?",
	}

	client := NewDefault()
	got, res, err := client.Issues.EditComment(context.Background(), "octocat/hello-world", 1, 1, input)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Comment)
	raw, _ := ioutil.ReadFile("testdata/issue_comment.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestIssueClose(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Patch("/repos/octocat/hello-world/issues/1").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/issue.json")

	client := NewDefault()
	res, err := client.Issues.Close(context.Background(), "octocat/hello-world", 1)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestIssueLock(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Put("/repos/octocat/hello-world/issues/1/lock").
		Reply(204).
		Type("application/json").
		SetHeaders(mockHeaders)

	client := NewDefault()
	res, err := client.Issues.Lock(context.Background(), "octocat/hello-world", 1)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestIssueUnlock(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Delete("/repos/octocat/hello-world/issues/1/lock").
		Reply(204).
		Type("application/json").
		SetHeaders(mockHeaders)

	client := NewDefault()
	res, err := client.Issues.Unlock(context.Background(), "octocat/hello-world", 1)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}
