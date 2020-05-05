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
// issue sub-tests
//

func testIssues(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		t.Run("List", testIssueList(client))
		t.Run("Find", testIssueFind(client))
		t.Run("Comments", testIssueComments(client))
	}
}

func testIssueList(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		opts := scm.IssueListOptions{
			Open:   true,
			Closed: true,
		}
		result, _, err := client.Issues.List(context.Background(), "gitlab-org/testme", opts)
		if err != nil {
			t.Error(err)
		}
		if len(result) == 0 {
			t.Errorf("Got empty issue list")
		}
	}
}

func testIssueFind(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		result, _, err := client.Issues.Find(context.Background(), "gitlab-org/testme", 1)
		if err != nil {
			t.Error(err)
		}
		t.Run("Issue", testIssue(result))
	}
}

//
// issue comment sub-tests
//

func testIssueComments(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Run("List", testIssueCommentList(client))
		t.Run("Find", testIssueCommentFind(client))
	}
}

func testIssueCommentList(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		opts := scm.ListOptions{}
		result, _, err := client.Issues.ListComments(context.Background(), "gitlab-org/testme", 1, opts)
		if err != nil {
			t.Error(err)
		}
		if len(result) == 0 {
			t.Errorf("Want a non-empty issue comment list")
		}
		for _, comment := range result {
			if comment.ID == 138083 {
				t.Run("Comment", testIssueComment(comment))
			}
		}
	}
}

func testIssueCommentFind(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		result, _, err := client.Issues.FindComment(context.Background(), "gitlab-org/testme", 1, 138083)
		if err != nil {
			t.Error(err)
		}
		t.Run("Comment", testIssueComment(result))
	}
}

//
// struct sub-tests
//

func testIssue(issue *scm.Issue) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := issue.Number, 1; got != want {
			t.Errorf("Want issue Number %d, got %d", want, got)
		}
		if got, want := issue.Title, "New issue"; got != want {
			t.Errorf("Want issue Title %q, got %q", want, got)
		}
		if got, want := issue.Link, "https://gitlab.com/gitlab-org/testme/-/issues/1"; got != want {
			t.Errorf("Want issue Title %q, got %q", want, got)
		}
		if got, want := issue.Author.Login, "marin"; got != want {
			t.Errorf("Want issue Author Login %q, got %q", want, got)
		}
		if got, want := issue.Author.Name, "Marin Jankovski"; got != want {
			t.Errorf("Want issue Author Name %q, got %q", want, got)
		}
		if got, want := issue.Author.Avatar, "https://secure.gravatar.com/avatar/5154f0b0eda3f798c8a254962d57192f?s=80&d=identicon"; got != want {
			t.Errorf("Want issue Author Name %q, got %q", want, got)
		}
		if got, want := issue.Closed, true; got != want {
			t.Errorf("Want issue Closed %v, got %v", want, got)
		}
		if got, want := issue.Created.Unix(), int64(1403255834); got != want {
			t.Errorf("Want issue Created %d, got %d", want, got)
		}
		if got, want := issue.Updated.Unix(), int64(1500004986); got != want {
			t.Errorf("Want issue Updated %d, got %d", want, got)
		}
	}
}

func testIssueComment(comment *scm.Comment) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := comment.ID, 138083; got != want {
			t.Errorf("Want issue comment ID %d, got %d", want, got)
		}
		if got, want := comment.Body, "Sorry, had to test something!"; got != want {
			t.Errorf("Want issue comment Body %q, got %q", want, got)
		}
		if got, want := comment.Author.Login, "jvanbaarsen"; got != want {
			t.Errorf("Want issue comment Author Login %q, got %q", want, got)
		}
		if got, want := comment.Author.Name, "Jeroen van Baarsen"; got != want {
			t.Errorf("Want issue comment Author Name %q, got %q", want, got)
		}
		if got, want := comment.Author.Avatar, "https://assets.gitlab-static.net/uploads/-/system/user/avatar/1164/avatar.png"; got != want {
			t.Errorf("Want issue comment Author Name %q, got %q", want, got)
		}
		if got, want := comment.Created.Unix(), int64(1403441879); got != want {
			t.Errorf("Want issue comment Created %d, got %d", want, got)
		}
		if got, want := comment.Updated.Unix(), int64(1403441879); got != want {
			t.Errorf("Want issue comment Updated %d, got %d", want, got)
		}
	}
}
