// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gitlab

import (
	"context"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
)

func TestReviewFind(t *testing.T) {
	service := new(reviewService)
	_, _, err := service.Find(context.Background(), "diaspora/diaspora", 1, 1)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestReviewList(t *testing.T) {
	service := new(reviewService)
	_, _, err := service.List(context.Background(), "diaspora/diaspora", 1, scm.ListOptions{})
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestReviewCreate(t *testing.T) {
	service := new(reviewService)
	_, _, err := service.Create(context.Background(), "diaspora/diaspora", 1, nil)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestReviewDelete(t *testing.T) {
	service := new(reviewService)
	_, err := service.Delete(context.Background(), "diaspora/diaspora", 1, 1)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}
