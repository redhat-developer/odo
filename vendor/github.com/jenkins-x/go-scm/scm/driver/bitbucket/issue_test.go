// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitbucket

import (
	"context"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
)

func TestIssueFind(t *testing.T) {
	_, _, err := NewDefault().Issues.Find(context.Background(), "", 0)
	if err != nil && err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestIssueCommentFind(t *testing.T) {
	_, _, err := NewDefault().Issues.FindComment(context.Background(), "", 0, 0)
	if err != nil && err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestIssueList(t *testing.T) {
	_, _, err := NewDefault().Issues.List(context.Background(), "", scm.IssueListOptions{})
	if err != nil && err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestIssueListComments(t *testing.T) {
	_, _, err := NewDefault().Issues.ListComments(context.Background(), "", 0, scm.ListOptions{})
	if err != nil && err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestIssueCreate(t *testing.T) {
	_, _, err := NewDefault().Issues.Create(context.Background(), "", &scm.IssueInput{})
	if err != nil && err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestIssueCreateComment(t *testing.T) {
	_, _, err := NewDefault().Issues.CreateComment(context.Background(), "", 0, &scm.CommentInput{})
	if err != nil && err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestIssueCommentDelete(t *testing.T) {
	_, err := NewDefault().Issues.DeleteComment(context.Background(), "", 0, 0)
	if err != nil && err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestIssueClose(t *testing.T) {
	_, err := NewDefault().Issues.Close(context.Background(), "", 0)
	if err != nil && err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestIssueLock(t *testing.T) {
	_, err := NewDefault().Issues.Lock(context.Background(), "", 0)
	if err != nil && err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestIssueUnlock(t *testing.T) {
	_, err := NewDefault().Issues.Unlock(context.Background(), "", 0)
	if err != nil && err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}
