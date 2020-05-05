// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stash

import (
	"context"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
)

func TestOrganizationFind(t *testing.T) {
	client, _ := New("https://api.bitbucket.org")
	_, _, err := client.Organizations.Find(context.Background(), "atlassian")
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestOrganizationList(t *testing.T) {
	client, _ := New("https://api.bitbucket.org")
	_, _, err := client.Organizations.List(context.Background(), scm.ListOptions{Size: 30, Page: 1})
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}
