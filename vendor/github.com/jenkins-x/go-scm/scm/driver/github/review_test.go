// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/h2non/gock"
	"github.com/jenkins-x/go-scm/scm"
)

func TestReviewFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/repos/octocat/hello-world/pulls/comments/1").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/pr_comment.json")

	client := NewDefault()
	got, res, err := client.Reviews.Find(context.Background(), "octocat/hello-world", 2, 1)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Review)
	raw, _ := ioutil.ReadFile("testdata/pr_comment.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestReviewList(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/repos/octocat/hello-world/pulls/1/comments").
		MatchParam("page", "1").
		MatchParam("per_page", "30").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/pr_comments.json")

	client := NewDefault()
	got, res, err := client.Reviews.List(context.Background(), "octocat/hello-world", 1, scm.ListOptions{Page: 1, Size: 30})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Review{}
	raw, _ := ioutil.ReadFile("testdata/pr_comments.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
	t.Run("Page", testPage(res))
}

func TestReviewCreate(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Post("/repos/octocat/hello-world/pulls/1/comments").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/pr_comment.json")

	input := &scm.ReviewInput{
		Body: "what?",
		Line: 1,
		Path: "file1.txt",
		Sha:  "6dcb09b5b57875f334f61aebed695e2e4193db5e",
	}

	client := NewDefault()
	got, res, err := client.Reviews.Create(context.Background(), "octocat/hello-world", 1, input)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Review)
	raw, _ := ioutil.ReadFile("testdata/pr_comment.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestReviewDelete(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Delete("/repos/octocat/hello-world/pulls/comments/1").
		Reply(204).
		Type("application/json").
		SetHeaders(mockHeaders)

	client := NewDefault()
	res, err := client.Reviews.Delete(context.Background(), "octocat/hello-world", 2, 1)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}
