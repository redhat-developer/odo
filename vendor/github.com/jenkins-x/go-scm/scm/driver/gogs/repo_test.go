// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gogs

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/h2non/gock"
	"github.com/jenkins-x/go-scm/scm"
)

//
// repository sub-tests
//

func TestRepoFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://try.gogs.io").
		Get("/api/v1/repos/gogits/gogs").
		Reply(200).
		Type("application/json").
		File("testdata/repo.json")

	client, _ := New("https://try.gogs.io")
	got, _, err := client.Repositories.Find(context.Background(), "gogits/gogs")
	if err != nil {
		t.Error(err)
	}

	want := new(scm.Repository)
	raw, _ := ioutil.ReadFile("testdata/repo.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestRepoFindPerm(t *testing.T) {
	defer gock.Off()

	gock.New("https://try.gogs.io").
		Get("/api/v1/repos/gogits/gogs").
		Reply(200).
		Type("application/json").
		File("testdata/repo.json")

	client, _ := New("https://try.gogs.io")
	got, _, err := client.Repositories.FindPerms(context.Background(), "gogits/gogs")
	if err != nil {
		t.Error(err)
	}

	want := new(scm.Repository)
	raw, _ := ioutil.ReadFile("testdata/repo.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want.Perm); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestRepoList(t *testing.T) {
	defer gock.Off()

	gock.New("https://try.gogs.io").
		Get("/api/v1/user/repos").
		Reply(200).
		Type("application/json").
		File("testdata/repos.json")

	client, _ := New("https://try.gogs.io")
	got, _, err := client.Repositories.List(context.Background(), scm.ListOptions{})
	if err != nil {
		t.Error(err)
	}

	want := []*scm.Repository{}
	raw, _ := ioutil.ReadFile("testdata/repos.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestRepoNotFound(t *testing.T) {
	defer gock.Off()

	gock.New("https://try.gogs.io").
		Get("/api/v1/repos/gogits/go-gogs-client").
		Reply(404).
		Type("text/plain")

	client, _ := New("https://try.gogs.io")
	_, _, err := client.Repositories.FindPerms(context.Background(), "gogits/go-gogs-client")
	if err == nil {
		t.Errorf("Expect Not Found error")
	} else if got, want := err.Error(), "Not Found"; got != want {
		t.Errorf("Want error %q, got %q", want, got)
	}
}

//
// hook sub-tests
//

func TestHookFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://try.gogs.io").
		Get("/api/v1/repos/gogits/gogs/hooks/20").
		Reply(200).
		Type("application/json").
		File("testdata/hook.json")

	client, _ := New("https://try.gogs.io")
	got, _, err := client.Repositories.FindHook(context.Background(), "gogits/gogs", "20")
	if err != nil {
		t.Error(err)
	}

	want := new(scm.Hook)
	raw, _ := ioutil.ReadFile("testdata/hook.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestHookList(t *testing.T) {
	defer gock.Off()

	gock.New("https://try.gogs.io").
		Get("/api/v1/repos/gogits/gogs/hooks").
		Reply(200).
		Type("application/json").
		File("testdata/hooks.json")

	client, _ := New("https://try.gogs.io")
	got, _, err := client.Repositories.ListHooks(context.Background(), "gogits/gogs", scm.ListOptions{})
	if err != nil {
		t.Error(err)
	}

	want := []*scm.Hook{}
	raw, _ := ioutil.ReadFile("testdata/hooks.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestHookCreate(t *testing.T) {
	defer gock.Off()

	gock.New("https://try.gogs.io").
		Post("/api/v1/repos/gogits/gogs/hooks").
		Reply(201).
		Type("application/json").
		File("testdata/hook.json")

	client, _ := New("https://try.gogs.io")
	got, _, err := client.Repositories.CreateHook(context.Background(), "gogits/gogs", &scm.HookInput{})
	if err != nil {
		t.Error(err)
	}

	want := new(scm.Hook)
	raw, _ := ioutil.ReadFile("testdata/hook.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestHookDelete(t *testing.T) {
	defer gock.Off()

	gock.New("https://try.gogs.io").
		Delete("/api/v1/repos/gogits/gogs/hooks/20").
		Reply(204).
		Type("application/json")

	client, _ := New("https://try.gogs.io")
	_, err := client.Repositories.DeleteHook(context.Background(), "gogits/gogs", "20")
	if err != nil {
		t.Error(err)
	}
}

func TestHookEvents(t *testing.T) {
	tests := []struct {
		in  scm.HookEvents
		out []string
	}{
		{
			in:  scm.HookEvents{Push: true},
			out: []string{"push"},
		},
		{
			in:  scm.HookEvents{Branch: true},
			out: []string{"create", "delete"},
		},
		{
			in:  scm.HookEvents{IssueComment: true},
			out: []string{"issue_comment"},
		},
		{
			in:  scm.HookEvents{PullRequestComment: true},
			out: []string{"issue_comment"},
		},
		{
			in:  scm.HookEvents{Issue: true},
			out: []string{"issues"},
		},
		{
			in:  scm.HookEvents{PullRequest: true},
			out: []string{"pull_request"},
		},
		{
			in: scm.HookEvents{
				Branch:             true,
				Issue:              true,
				IssueComment:       true,
				PullRequest:        true,
				PullRequestComment: true,
				Push:               true,
				ReviewComment:      true,
				Tag:                true,
			},
			out: []string{"pull_request", "issues", "issue_comment", "create", "delete", "push"},
		},
	}
	for _, test := range tests {
		got, want := convertHookEvent(test.in), test.out
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("Unexpected Results")
			t.Log(diff)
		}
	}
}

//
// status sub-tests
//

func TestStatusList(t *testing.T) {
	client, _ := New("https://try.gogs.io")
	_, _, err := client.Repositories.ListStatus(context.Background(), "gogits/gogs", "master", scm.ListOptions{})
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestStatusCreate(t *testing.T) {
	client, _ := New("https://try.gogs.io")
	_, _, err := client.Repositories.CreateStatus(context.Background(), "gogits/gogs", "master", &scm.StatusInput{})
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}
