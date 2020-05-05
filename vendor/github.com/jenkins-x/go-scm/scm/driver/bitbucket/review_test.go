// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitbucket

import (
	"context"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
)

func TestReviewFind(t *testing.T) {
	_, _, err := NewDefault().Reviews.Find(context.Background(), "", 0, 0)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestReviewList(t *testing.T) {
	_, _, err := NewDefault().Reviews.List(context.Background(), "", 0, scm.ListOptions{})
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestReviewCreate(t *testing.T) {
	_, _, err := NewDefault().Reviews.Create(context.Background(), "", 0, &scm.ReviewInput{})
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestReviewDelete(t *testing.T) {
	_, err := NewDefault().Reviews.Delete(context.Background(), "", 0, 0)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}
