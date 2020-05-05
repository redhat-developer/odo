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
		result, _, err := client.Issues.List(context.Background(), "octocat/Hello-World", opts)
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
		result, _, err := client.Issues.Find(context.Background(), "octocat/Hello-World", 348)
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
		result, _, err := client.Issues.ListComments(context.Background(), "octocat/Hello-World", 348, opts)
		if err != nil {
			t.Error(err)
		}
		if len(result) == 0 {
			t.Errorf("Want a non-empty issue comment list")
		}
		for _, comment := range result {
			if comment.ID == 304068667 {
				t.Run("Comment", testIssueComment(comment))
			}
		}
	}
}

func testIssueCommentFind(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		result, _, err := client.Issues.FindComment(context.Background(), "octocat/Hello-World", 348, 304068667)
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
		if got, want := issue.Number, 348; got != want {
			t.Errorf("Want issue Number %d, got %d", want, got)
		}
		if got, want := issue.Title, "Testing comments"; got != want {
			t.Errorf("Want issue Title %q, got %q", want, got)
		}
		if got, want := issue.Body, "Let's add some, shall we?"; got != want {
			t.Errorf("Want issue Body %q, got %q", want, got)
		}
		if got, want := issue.Link, "https://github.com/octocat/Hello-World/issues/348"; got != want {
			t.Errorf("Want issue Link %q, got %q", want, got)
		}
		if got, want := issue.Author.Login, "octocat"; got != want {
			t.Errorf("Want issue Author Login %q, got %q", want, got)
		}
		if got, want := issue.Author.Avatar, "https://avatars3.githubusercontent.com/u/583231?v=4"; got != want {
			t.Errorf("Want issue Author Name %q, got %q", want, got)
		}
		if got, want := issue.Closed, false; got != want {
			t.Errorf("Want issue Closed %v, got %v", want, got)
		}
		if got, want := issue.Created.Unix(), int64(1495478858); got != want {
			t.Errorf("Want issue Created %d, got %d", want, got)
		}
	}
}

func testIssueComment(comment *scm.Comment) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := comment.ID, 304068667; got != want {
			t.Errorf("Want issue comment ID %d, got %d", want, got)
		}
		if got, want := comment.Body, "A shiny new comment! :tada:"; got != want {
			t.Errorf("Want issue comment Body %q, got %q", want, got)
		}
		if got, want := comment.Author.Login, "defualt"; got != want {
			t.Errorf("Want issue comment Author Login %q, got %q", want, got)
		}
		// TODO: Avatar check seems to have become unreliable. Reenable in the future. (apb)
		//if got, want := comment.Author.Avatar, "https://avatars2.githubusercontent.com/u/399135?v=4"; got != want {
		//	t.Errorf("Want issue comment Author Name %q, got %q", want, got)
		//}
		if got, want := comment.Created.Unix(), int64(1495732818); got != want {
			t.Errorf("Want issue comment Created %d, got %d", want, got)
		}
		if got, want := comment.Updated.Unix(), int64(1495732818); got != want {
			t.Errorf("Want issue comment Updated %d, got %d", want, got)
		}
	}
}
