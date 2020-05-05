// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitbucket

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/jenkins-x/go-scm/scm"

	"github.com/google/go-cmp/cmp"
	"github.com/h2non/gock"
)

func TestUserFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Get("/2.0/user").
		Reply(200).
		Type("application/json").
		File("testdata/user.json")

	client, _ := New("https://api.bitbucket.org")
	got, _, err := client.Users.Find(context.Background())
	if err != nil {
		t.Error(err)
	}

	want := new(scm.User)
	raw, _ := ioutil.ReadFile("testdata/user.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestUserLoginFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Get("/2.0/users/brydzewski").
		Reply(200).
		Type("application/json").
		File("testdata/user.json")

	client, _ := New("https://api.bitbucket.org")
	got, _, err := client.Users.FindLogin(context.Background(), "brydzewski")
	if err != nil {
		t.Error(err)
	}

	want := new(scm.User)
	raw, _ := ioutil.ReadFile("testdata/user.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestUserFindEmail(t *testing.T) {
	client, _ := New("https://api.bitbucket.org")
	_, _, err := client.Users.FindEmail(context.Background())
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}
