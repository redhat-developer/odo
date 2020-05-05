// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gitea

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

	gock.New("https://try.gitea.io").
		Get("/api/v1/user").
		Reply(200).
		Type("application/json").
		File("testdata/user.json")

	client, _ := New("https://try.gitea.io")
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

	gock.New("https://try.gitea.io").
		Get("/api/v1/users/jcitizen").
		Reply(200).
		Type("application/json").
		File("testdata/user.json")

	client, _ := New("https://try.gitea.io")
	got, _, err := client.Users.FindLogin(context.Background(), "jcitizen")
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
	defer gock.Off()

	gock.New("https://try.gitea.io").
		Get("/api/v1/user").
		Reply(200).
		Type("application/json").
		File("testdata/user.json")

	client, _ := New("https://try.gitea.io")
	email, _, err := client.Users.FindEmail(context.Background())
	if err != nil {
		t.Error(err)
	}

	if got, want := email, "jane@example.com"; got != want {
		t.Errorf("Want email %s, got %s", want, got)
	}
}
