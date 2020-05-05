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
// issue sub-tests
//

func TestIssueFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://try.gitea.io").
		Get("/api/v1/repos/go-gitea/gitea/issues/1").
		Reply(200).
		Type("application/json").
		File("testdata/issue.json")

	client, _ := New("https://try.gitea.io")
	got, _, err := client.Issues.Find(context.Background(), "go-gitea/gitea", 1)
	if err != nil {
		t.Error(err)
	}

	want := new(scm.Issue)
	raw, _ := ioutil.ReadFile("testdata/issue.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestIssueList(t *testing.T) {
	defer gock.Off()

	gock.New("https://try.gitea.io").
		Get("/api/v1/repos/go-gitea/gitea/issues").
		Reply(200).
		Type("application/json").
		File("testdata/issues.json")

	client, _ := New("https://try.gitea.io")
	got, _, err := client.Issues.List(context.Background(), "go-gitea/gitea", scm.IssueListOptions{})
	if err != nil {
		t.Error(err)
	}

	want := []*scm.Issue{}
	raw, _ := ioutil.ReadFile("testdata/issues.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestIssueCreate(t *testing.T) {
	defer gock.Off()

	gock.New("https://try.gitea.io").
		Post("/api/v1/repos/go-gitea/gitea/issues").
		Reply(200).
		Type("application/json").
		File("testdata/issue.json")

	input := scm.IssueInput{
		Title: "Bug found",
		Body:  "I'm having a problem with this.",
	}

	client, _ := New("https://try.gitea.io")
	got, _, err := client.Issues.Create(context.Background(), "go-gitea/gitea", &input)
	if err != nil {
		t.Error(err)
	}

	want := new(scm.Issue)
	raw, _ := ioutil.ReadFile("testdata/issue.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestIssueClose(t *testing.T) {
	client, _ := New("https://try.gitea.io")
	_, err := client.Issues.Close(context.Background(), "gogits/go-gogs-client", 1)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestIssueLock(t *testing.T) {
	client, _ := New("https://try.gitea.io")
	_, err := client.Issues.Lock(context.Background(), "gogits/go-gogs-client", 1)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestIssueUnlock(t *testing.T) {
	client, _ := New("https://try.gitea.io")
	_, err := client.Issues.Unlock(context.Background(), "gogits/go-gogs-client", 1)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

//
// issue comment sub-tests
//

func TestIssueCommentFind(t *testing.T) {
	client, _ := New("https://try.gitea.io")
	_, _, err := client.Issues.FindComment(context.Background(), "gogits/go-gogs-client", 1, 1)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestIssueCommentList(t *testing.T) {
	defer gock.Off()

	gock.New("https://try.gitea.io").
		Get("/api/v1/repos/go-gitea/gitea/issues/1/comments").
		Reply(200).
		Type("application/json").
		File("testdata/comments.json")

	client, _ := New("https://try.gitea.io")
	got, _, err := client.Issues.ListComments(context.Background(), "go-gitea/gitea", 1, scm.ListOptions{})
	if err != nil {
		t.Error(err)
	}

	want := []*scm.Comment{}
	raw, _ := ioutil.ReadFile("testdata/comments.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestIssueCommentCreate(t *testing.T) {
	defer gock.Off()

	gock.New("https://try.gitea.io").
		Post("/api/v1/repos/go-gitea/gitea/issues/1/comments").
		Reply(201).
		Type("application/json").
		File("testdata/comment.json")

	client, _ := New("https://try.gitea.io")
	got, _, err := client.Issues.CreateComment(context.Background(), "go-gitea/gitea", 1, &scm.CommentInput{Body: "what?"})
	if err != nil {
		t.Error(err)
	}

	want := new(scm.Comment)
	raw, _ := ioutil.ReadFile("testdata/comment.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	if gock.IsPending() {
		t.Errorf("Pending API calls")
	}
}

func TestIssueCommentDelete(t *testing.T) {
	defer gock.Off()

	gock.New("https://try.gitea.io").
		Delete("/api/v1/repos/go-gitea/gitea/issues/1/comments/1").
		Reply(204).
		Type("application/json")

	client, _ := New("https://try.gitea.io")
	_, err := client.Issues.DeleteComment(context.Background(), "go-gitea/gitea", 1, 1)
	if err != nil {
		t.Error(err)
	}
}
