// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gitea

import (
	"context"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
)

func TestReviewFind(t *testing.T) {
	client, _ := New("https://try.gitea.io")
	_, _, err := client.Reviews.Find(context.Background(), "go-gitea/gitea", 1, 1)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestReviewList(t *testing.T) {
	client, _ := New("https://try.gitea.io")
	_, _, err := client.Reviews.List(context.Background(), "go-gitea/gitea", 1, scm.ListOptions{})
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestReviewCreate(t *testing.T) {
	client, _ := New("https://try.gitea.io")
	_, _, err := client.Reviews.Create(context.Background(), "go-gitea/gitea", 1, nil)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestReviewDelete(t *testing.T) {
	client, _ := New("https://try.gitea.io")
	_, err := client.Reviews.Delete(context.Background(), "go-gitea/gitea", 1, 1)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}
