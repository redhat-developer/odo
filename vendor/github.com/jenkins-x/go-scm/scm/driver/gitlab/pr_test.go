// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gitlab

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/h2non/gock"
	"github.com/jenkins-x/go-scm/scm"
)

func TestPullFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://gitlab.com").
		Get("/api/v4/projects/diaspora/diaspora/merge_requests/1347").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/merge.json")

	client := NewDefault()
	got, res, err := client.PullRequests.Find(context.Background(), "diaspora/diaspora", 1347)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.PullRequest)
	raw, err := ioutil.ReadFile("testdata/merge.json.golden")
	if err != nil {
		t.Fatalf("ioutil.ReadFile: %v", err)
	}
	if err := json.Unmarshal(raw, want); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestPullList(t *testing.T) {
	defer gock.Off()

	updatedAfter, _ := time.Parse(scm.SearchTimeFormat, "2015-12-18T17:30:22.522Z")
	gock.New("https://gitlab.com").
		Get("/api/v4/projects/diaspora/diaspora/merge_requests").
		MatchParam("labels", "Community contribution,Manage").
		MatchParam("page", "1").
		MatchParam("per_page", "30").
		MatchParam("state", "all").
		MatchParam("updated_after", "2015-12-18T17:30:22Z").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/merges.json")

	client := NewDefault()
	got, res, err := client.PullRequests.List(context.Background(), "diaspora/diaspora", scm.PullRequestListOptions{
		Page:         1,
		Size:         30,
		Open:         true,
		Closed:       true,
		Labels:       []string{"Community contribution", "Manage"},
		UpdatedAfter: &updatedAfter,
	})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.PullRequest{}
	raw, err := ioutil.ReadFile("testdata/merges.json.golden")
	if err != nil {
		t.Fatalf("ioutil.ReadFile: %v", err)
	}
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
	t.Run("Page", testPage(res))
}

func TestPullListChanges(t *testing.T) {
	gock.New("https://gitlab.com").
		Get("/api/v4/projects/diaspora/diaspora/merge_requests/1347/changes").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/merge_diff.json")

	client := NewDefault()
	got, res, err := client.PullRequests.ListChanges(context.Background(), "diaspora/diaspora", 1347, scm.ListOptions{Page: 1, Size: 30})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Change{}
	raw, _ := ioutil.ReadFile("testdata/merge_diff.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestPullMerge(t *testing.T) {
	defer gock.Off()

	gock.New("https://gitlab.com").
		Put("/api/v4/projects/diaspora/diaspora/merge_requests/1347/merge").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders)

	client := NewDefault()
	res, err := client.PullRequests.Merge(context.Background(), "diaspora/diaspora", 1347, nil)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestPullClose(t *testing.T) {
	defer gock.Off()

	gock.New("https://gitlab.com").
		Put("/api/v4/projects/diaspora/diaspora/merge_requests/1347").
		MatchParam("state_event", "closed").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders)

	client := NewDefault()
	res, err := client.PullRequests.Close(context.Background(), "diaspora/diaspora", 1347)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestPullCommentFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://gitlab.com").
		Get("/api/v4/projects/diaspora/diaspora/merge_requests/2/notes/1").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/merge_note.json")

	client := NewDefault()
	got, res, err := client.PullRequests.FindComment(context.Background(), "diaspora/diaspora", 2, 1)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Comment)
	raw, _ := ioutil.ReadFile("testdata/merge_note.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestPullListComments(t *testing.T) {
	defer gock.Off()

	gock.New("https://gitlab.com").
		Get("/api/v4/projects/diaspora/diaspora/merge_requests/1/notes").
		MatchParam("page", "1").
		MatchParam("per_page", "30").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/merge_notes.json")

	client := NewDefault()
	got, res, err := client.PullRequests.ListComments(context.Background(), "diaspora/diaspora", 1, scm.ListOptions{Size: 30, Page: 1})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Comment{}
	raw, _ := ioutil.ReadFile("testdata/merge_notes.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
	t.Run("Page", testPage(res))
}

func TestPullCreateComment(t *testing.T) {
	defer gock.Off()

	gock.New("https://gitlab.com").
		Post("/api/v4/projects/diaspora/diaspora/merge_requests/1/notes").
		MatchParam("body", "lgtm").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/merge_note.json")

	input := &scm.CommentInput{
		Body: "lgtm",
	}

	client := NewDefault()
	got, res, err := client.PullRequests.CreateComment(context.Background(), "diaspora/diaspora", 1, input)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Comment)
	raw, _ := ioutil.ReadFile("testdata/merge_note.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestPullCommentDelete(t *testing.T) {
	defer gock.Off()

	gock.New("https://gitlab.com").
		Delete("/api/v4/projects/diaspora/diaspora/merge_requests/2/notes/1").
		Reply(204).
		Type("application/json").
		SetHeaders(mockHeaders)

	client := NewDefault()
	res, err := client.PullRequests.DeleteComment(context.Background(), "diaspora/diaspora", 2, 1)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestPullEditComment(t *testing.T) {
	defer gock.Off()

	gock.New("https://gitlab.com").
		Put("/api/v4/projects/diaspora/diaspora/merge_requests/2/notes/1").
		File("testdata/edit_issue_note.json").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/merge_note.json")

	input := &scm.CommentInput{
		Body: "closed",
	}

	client := NewDefault()
	got, res, err := client.PullRequests.EditComment(context.Background(), "diaspora/diaspora", 2, 1, input)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Comment)
	raw, _ := ioutil.ReadFile("testdata/merge_note.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestPullCreate(t *testing.T) {
	defer gock.Off()

	gock.New("https://gitlab.com").
		Post("/api/v4/projects/diaspora/diaspora/merge_requests").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/pr_create.json")

	input := &scm.PullRequestInput{
		Title: "Amazing new feature",
		Body:  "Please pull these awesome changes in!",
		Head:  "test1",
		Base:  "master",
	}

	client := NewDefault()
	got, res, err := client.PullRequests.Create(context.Background(), "diaspora/diaspora", input)
	if err != nil {
		t.Fatal(err)
	}

	want := new(scm.PullRequest)
	raw, _ := ioutil.ReadFile("testdata/pr_create.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestPullListEvents(t *testing.T) {
	defer gock.Off()

	gock.New("https://gitlab.com").
		Get("/api/v4/projects/diaspora/diaspora/merge_requests/28/resource_label_events").
		MatchParam("page", "1").
		MatchParam("per_page", "30").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/pr_events.json")

	client := NewDefault()
	got, res, err := client.PullRequests.ListEvents(context.Background(), "diaspora/diaspora", 28, scm.ListOptions{Size: 30, Page: 1})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.ListedIssueEvent{}
	raw, _ := ioutil.ReadFile("testdata/pr_events.golden.json")
	err = json.Unmarshal(raw, &want)
	if err != nil {
		t.Error(err)
		return
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
	t.Run("Page", testPage(res))
}
