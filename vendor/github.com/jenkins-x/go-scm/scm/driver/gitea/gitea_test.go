// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package gitea implements a Gogs client.
package gitea

import "testing"

func TestClient(t *testing.T) {
	client, err := New("https://try.gitea.io")
	if err != nil {
		t.Error(err)
	}
	if got, want := client.BaseURL.String(), "https://try.gitea.io/"; got != want {
		t.Errorf("Want Client URL %q, got %q", want, got)
	}
}

func TestClient_Base(t *testing.T) {
	client, err := New("https://try.gitea.io/v1")
	if err != nil {
		t.Error(err)
	}
	if got, want := client.BaseURL.String(), "https://try.gitea.io/v1/"; got != want {
		t.Errorf("Want Client URL %q, got %q", want, got)
	}
}

func TestClient_Error(t *testing.T) {
	_, err := New("http://a b.com/")
	if err == nil {
		t.Errorf("Expect error when invalid URL")
	}
}
