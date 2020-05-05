// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/jenkins-x/go-scm/scm"

	"github.com/google/go-cmp/cmp"
	"github.com/h2non/gock"
)

func TestUserFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/user").
		Reply(200).
		Type("application/json").
		SetHeader("X-GitHub-Request-Id", "DD0E:6011:12F21A8:1926790:5A2064E2").
		SetHeader("X-RateLimit-Limit", "60").
		SetHeader("X-RateLimit-Remaining", "59").
		SetHeader("X-RateLimit-Reset", "1512076018").
		File("testdata/user.json")

	client := NewDefault()
	got, res, err := client.Users.Find(context.Background())
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.User)
	raw, _ := ioutil.ReadFile("testdata/user.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestUserLoginFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/users/octocat").
		Reply(200).
		Type("application/json").
		SetHeader("X-GitHub-Request-Id", "DD0E:6011:12F21A8:1926790:5A2064E2").
		SetHeader("X-RateLimit-Limit", "60").
		SetHeader("X-RateLimit-Remaining", "59").
		SetHeader("X-RateLimit-Reset", "1512076018").
		File("testdata/user.json")

	client := NewDefault()
	got, res, err := client.Users.FindLogin(context.Background(), "octocat")
	if err != nil {
		t.Error(err)
	}

	want := new(scm.User)
	raw, _ := ioutil.ReadFile("testdata/user.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
		json.NewEncoder(os.Stdout).Encode(got)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestUserEmailFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/user").
		Reply(200).
		Type("application/json").
		SetHeader("X-GitHub-Request-Id", "DD0E:6011:12F21A8:1926790:5A2064E2").
		SetHeader("X-RateLimit-Limit", "60").
		SetHeader("X-RateLimit-Remaining", "59").
		SetHeader("X-RateLimit-Reset", "1512076018").
		File("testdata/user.json")

	client := NewDefault()
	result, res, err := client.Users.FindEmail(context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	if got, want := result, "octocat@github.com"; got != want {
		t.Errorf("Want user Email %q, got %q", want, got)
	}
	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}
