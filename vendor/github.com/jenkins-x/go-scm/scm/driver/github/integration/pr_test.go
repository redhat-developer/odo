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
// pull request sub-tests
//

func testPullRequests(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		t.Run("List", testPullRequestList(client))
		t.Run("Find", testPullRequestFind(client))
		t.Run("Changes", testPullRequestChanges(client))
		t.Run("Comments", testPullRequestComments(client))
	}
}

func testPullRequestList(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		opts := scm.PullRequestListOptions{
			Open:   true,
			Closed: true,
		}
		result, _, err := client.PullRequests.List(context.Background(), "octocat/Hello-World", opts)
		if err != nil {
			t.Error(err)
		}
		if len(result) == 0 {
			t.Errorf("Got empty pull request list")
		}
		for _, pr := range result {
			if pr.Number == 1 {
				t.Run("PullRequest", testPullRequest(pr))
			}
		}
	}
}

func testPullRequestFind(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		result, _, err := client.PullRequests.Find(context.Background(), "octocat/Hello-World", 140)
		if err != nil {
			t.Error(err)
		}
		t.Run("PullRequest", testPullRequest(result))
	}
}

//
// pull request comment sub-tests
//

func testPullRequestComments(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		t.Run("List", testPullRequestCommentFind(client))
		t.Run("Find", testPullRequestCommentList(client))
	}
}

func testPullRequestCommentFind(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		result, _, err := client.PullRequests.FindComment(context.Background(), "octocat/Hello-World", 140, 60475333)
		if err != nil {
			t.Error(err)
		}
		t.Run("Comment", testPullRequestComment(result))
	}
}

func testPullRequestCommentList(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		opts := scm.ListOptions{}
		result, _, err := client.PullRequests.ListComments(context.Background(), "octocat/Hello-World", 140, opts)
		if err != nil {
			t.Error(err)
		}
		if len(result) == 0 {
			t.Errorf("Got empty pull request comment list")
		}
		for _, comment := range result {
			if comment.ID == 2990882 {
				t.Run("Comment", testPullRequestComment(comment))
			}
		}
	}
}

//
// pull request changes sub-tests
//

func testPullRequestChanges(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		opts := scm.ListOptions{}
		result, _, err := client.PullRequests.ListChanges(context.Background(), "octocat/Hello-World", 140, opts)
		if err != nil {
			t.Error(err)
		}
		if len(result) == 0 {
			t.Errorf("Got empty pull request change list")
			return
		}
		t.Run("File", testChange(result[0]))
	}
}

//
// struct sub-tests
//

func testPullRequest(pr *scm.PullRequest) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := pr.Number, 140; got != want {
			t.Errorf("Want pr Number %d, got %d", want, got)
		}
		if got, want := pr.Title, "Create CONTRIBUTING.md"; got != want {
			t.Errorf("Want pr Title %q, got %q", want, got)
		}
		if got, want := pr.Body, ""; got != want {
			t.Errorf("Want pr Body %q, got %q", want, got)
		}
		if got, want := pr.Source, "test"; got != want {
			t.Errorf("Want pr Source %q, got %q", want, got)
		}
		if got, want := pr.Target, "master"; got != want {
			t.Errorf("Want pr Target %q, got %q", want, got)
		}
		if got, want := pr.Ref, "refs/pull/140/head"; got != want {
			t.Errorf("Want pr Ref %q, got %q", want, got)
		}
		if got, want := pr.Sha, "b3cbd5bbd7e81436d2eee04537ea2b4c0cad4cdf"; got != want {
			t.Errorf("Want pr Sha %q, got %q", want, got)
		}
		if got, want := pr.Link, "https://github.com/octocat/Hello-World/pull/140.diff"; got != want {
			t.Errorf("Want pr Link %q, got %q", want, got)
		}
		if got, want := pr.Author.Login, "octocat"; got != want {
			t.Errorf("Want pr Author Login %q, got %q", want, got)
		}
		if got, want := pr.Author.Avatar, "https://avatars3.githubusercontent.com/u/583231?v=4"; got != want {
			t.Errorf("Want pr Author Avatar %q, got %q", want, got)
		}
		if got, want := pr.Closed, true; got != want {
			t.Errorf("Want pr Closed %v, got %v", want, got)
		}
		if got, want := pr.Merged, false; got != want {
			t.Errorf("Want pr Merged %v, got %v", want, got)
		}
		if got, want := pr.Created.Unix(), int64(1402523517); got != want {
			t.Errorf("Want pr Created %d, got %d", want, got)
		}
		if got, want := pr.Updated.Unix(), int64(1414224441); got != want {
			t.Errorf("Want pr Updated %d, got %d", want, got)
		}
	}
}

func testPullRequestComment(comment *scm.Comment) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := comment.ID, 60475333; got != want {
			t.Errorf("Want pr comment ID %d, got %d", want, got)
		}
		if got, want := comment.Body, "wwwwwwwaa\n"; got != want {
			t.Errorf("Want pr comment Body %q, got %q", want, got)
		}
		if got, want := comment.Author.Login, "tompang"; got != want {
			t.Errorf("Want pr comment Author Login %q, got %q", want, got)
		}
		if got, want := comment.Author.Name, ""; got != want {
			t.Errorf("Want pr comment Author Name %q, got %q", want, got)
		}
		// TODO: Avatar check has become unreliable. Fix in the future. (apb)
		//if got, want := comment.Author.Avatar, "https://avatars3.githubusercontent.com/u/7744744?v=4"; got != want {
		//	t.Errorf("Want pr comment Author Avatar %q, got %q", want, got)
		//}
		if got, want := comment.Created.Unix(), int64(1414224391); got != want {
			t.Errorf("Want pr comment Created %d, got %d", want, got)
		}
		if got, want := comment.Updated.Unix(), int64(1414224407); got != want {
			t.Errorf("Want pr comment Updated %d, got %d", want, got)
		}
	}
}

func testChange(change *scm.Change) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := change.Path, "CONTRIBUTING.md"; got != want {
			t.Errorf("Want file change Path %q, got %q", want, got)
		}
		if got, want := change.Added, true; got != want {
			t.Errorf("Want file Added %v, got %v", want, got)
		}
		if got, want := change.Deleted, false; got != want {
			t.Errorf("Want file Deleted %v, got %v", want, got)
		}
		if got, want := change.Renamed, false; got != want {
			t.Errorf("Want file Renamed %v, got %v", want, got)
		}
	}
}
