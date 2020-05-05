// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gogs

import (
	"context"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
)

//
// pull request sub-tests
//

func TestPullRequestFind(t *testing.T) {
	client, _ := New("https://try.gogs.io")
	_, _, err := client.PullRequests.Find(context.Background(), "gogits/gogs", 1)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestPullRequestList(t *testing.T) {
	client, _ := New("https://try.gogs.io")
	_, _, err := client.PullRequests.List(context.Background(), "gogits/gogs", scm.PullRequestListOptions{})
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestPullRequestClose(t *testing.T) {
	client, _ := New("https://try.gogs.io")
	_, err := client.PullRequests.Close(context.Background(), "gogits/gogs", 1)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestPullRequestMerge(t *testing.T) {
	client, _ := New("https://try.gogs.io")
	_, err := client.PullRequests.Merge(context.Background(), "gogits/gogs", 1, nil)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

//
// pull request change sub-tests
//

func TestPullRequestChanges(t *testing.T) {
	client, _ := New("https://try.gogs.io")
	_, _, err := client.PullRequests.ListChanges(context.Background(), "gogits/gogs", 1, scm.ListOptions{})
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

//
// pull request comment sub-tests
//

func TestPullRequestCommentFind(t *testing.T) {
	client, _ := New("https://try.gogs.io")
	_, _, err := client.PullRequests.FindComment(context.Background(), "gogits/gogs", 1, 1)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestPullRequestCommentList(t *testing.T) {
	client, _ := New("https://try.gogs.io")
	_, _, err := client.PullRequests.ListComments(context.Background(), "gogits/gogs", 1, scm.ListOptions{})
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestPullRequestCommentCreate(t *testing.T) {
	client, _ := New("https://try.gogs.io")
	_, _, err := client.PullRequests.CreateComment(context.Background(), "gogits/gogs", 1, &scm.CommentInput{})
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestPullRequestCommentDelete(t *testing.T) {
	client, _ := New("https://try.gogs.io")
	_, err := client.PullRequests.DeleteComment(context.Background(), "gogits/gogs", 1, 1)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestPullCreate(t *testing.T) {
	client, _ := New("https://try.gogs.io")
	input := &scm.PullRequestInput{
		Title: "Gogs feature",
		Body:  "New Gogs feature",
		Head:  "new-feature",
		Base:  "master",
	}

	_, _, err := client.PullRequests.Create(context.Background(), "gogits/gogs", input)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}
