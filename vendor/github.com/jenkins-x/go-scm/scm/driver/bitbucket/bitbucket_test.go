// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitbucket

import (
	"testing"

	"github.com/jenkins-x/go-scm/scm"
)

func TestClient(t *testing.T) {
	client, err := New("https://api.bitbucket.org")
	if err != nil {
		t.Error(err)
	}
	if got, want := client.BaseURL.String(), "https://api.bitbucket.org/"; got != want {
		t.Errorf("Want Client URL %q, got %q", want, got)
	}
}

func TestClient_Base(t *testing.T) {
	client, err := New("https://api.bitbucket.org/v1")
	if err != nil {
		t.Error(err)
	}
	if got, want := client.BaseURL.String(), "https://api.bitbucket.org/v1/"; got != want {
		t.Errorf("Want Client URL %q, got %q", want, got)
	}
}

func TestClient_Default(t *testing.T) {
	client := NewDefault()
	if got, want := client.BaseURL.String(), "https://api.bitbucket.org/"; got != want {
		t.Errorf("Want Client URL %q, got %q", want, got)
	}
}

func TestClient_Error(t *testing.T) {
	_, err := New("http://a b.com/")
	if err == nil {
		t.Errorf("Expect error when invalid URL")
	}
}

func testPage(res *scm.Response) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := res.Page.Next, 2; got != want {
			t.Errorf("Want next page %d, got %d", want, got)
		}
		if got, want := res.Page.First, 1; got != want {
			t.Errorf("Want first page %d, got %d", want, got)
		}
	}
}
