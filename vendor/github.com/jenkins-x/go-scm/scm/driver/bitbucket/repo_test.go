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

func TestRepositoryFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Get("/2.0/repositories/atlassian/stash-example-plugin").
		Reply(200).
		Type("application/json").
		File("testdata/repo.json")

	client, _ := New("https://api.bitbucket.org")
	got, _, err := client.Repositories.Find(context.Background(), "atlassian/stash-example-plugin")
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

func TestRepositoryFind_NotFound(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Get("/2.0/repositories/dev/null").
		Reply(404).
		Type("application/json").
		File("testdata/error.json")

	client, _ := New("https://api.bitbucket.org")
	_, _, err := client.Repositories.Find(context.Background(), "dev/null")
	if err == nil {
		t.Errorf("Expect not found message")
	}

	if got, want := err.Error(), "Repository dev/null not found"; got != want {
		t.Errorf("Want error message %q, got %q", want, got)
	}
}

func TestRepositoryPerms(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Get("/2.0/user/permissions/repositories").
		// MatchParam("repository.full_name", `"atlassian/stash-example-plugin"`).
		Reply(200).
		Type("application/json").
		File("testdata/perms.json")

	client, _ := New("https://api.bitbucket.org")
	got, _, err := client.Repositories.FindPerms(context.Background(), "atlassian/stash-example-plugin")
	if err != nil {
		t.Error(err)
	}

	want := new(scm.Perm)
	raw, _ := ioutil.ReadFile("testdata/perms.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestRepositoryList(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Get("/2.0/repositories").
		MatchParam("after", "PLACEHOLDER").
		MatchParam("pagelen", "1").
		MatchParam("role", "member").
		Reply(200).
		Type("application/json").
		File("testdata/repos-2.json")

	gock.New("https://api.bitbucket.org").
		Get("/2.0/repositories").
		MatchParam("pagelen", "1").
		MatchParam("role", "member").
		Reply(200).
		Type("application/json").
		File("testdata/repos.json")

	got := []*scm.Repository{}
	opts := scm.ListOptions{Size: 1}
	client, _ := New("https://api.bitbucket.org")

	for {
		ctx := context.Background()
		repos, res, err := client.Repositories.List(ctx, opts)
		if err != nil {
			t.Error(err)
		}
		got = append(got, repos...)

		opts.Page = res.Page.Next
		opts.URL = res.Page.NextURL

		if opts.Page == 0 && opts.URL == "" {
			break
		}
	}

	want := []*scm.Repository{}
	raw, _ := ioutil.ReadFile("testdata/repos.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestStatusList(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Get("/2.0/repositories/atlassian/stash-example-plugin/commit/a6e5e7d797edf751cbd839d6bd4aef86c941eec9/statuses").
		MatchParam("page", "1").
		MatchParam("pagelen", "30").
		Reply(200).
		Type("application/json").
		File("testdata/statuses.json")

	client, _ := New("https://api.bitbucket.org")
	got, res, err := client.Repositories.ListStatus(context.Background(), "atlassian/stash-example-plugin", "a6e5e7d797edf751cbd839d6bd4aef86c941eec9", scm.ListOptions{Size: 30, Page: 1})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Status{}
	raw, _ := ioutil.ReadFile("testdata/statuses.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Page", testPage(res))
}

func TestStatusCreate(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Post("/2.0/repositories/atlassian/stash-example-plugin/commit/a6e5e7d797edf751cbd839d6bd4aef86c941eec9/statuses/build").
		File("testdata/status_input.json").
		Reply(201).
		Type("application/json").
		File("testdata/status.json")

	in := &scm.StatusInput{
		Desc:   "Build has completed successfully",
		Label:  "continuous-integration/drone",
		State:  scm.StateSuccess,
		Target: "https://ci.example.com/1000/output",
	}

	client, _ := New("https://api.bitbucket.org")
	got, _, err := client.Repositories.CreateStatus(context.Background(), "atlassian/stash-example-plugin", "a6e5e7d797edf751cbd839d6bd4aef86c941eec9", in)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Status)
	raw, _ := ioutil.ReadFile("testdata/status.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestRepositoryHookFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Get("/2.0/repositories/atlassian/stash-example-plugin/hooks/{d53603cc-3f67-45ea-b310-aaa5ef6ec061}").
		Reply(200).
		Type("application/json").
		File("testdata/hook.json")

	client, _ := New("https://api.bitbucket.org")
	got, _, err := client.Repositories.FindHook(context.Background(), "atlassian/stash-example-plugin", "{d53603cc-3f67-45ea-b310-aaa5ef6ec061}")
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Hook)
	raw, _ := ioutil.ReadFile("testdata/hook.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestRepositoryHookList(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Get("/2.0/repositories/atlassian/stash-example-plugin/hooks").
		MatchParam("page", "1").
		MatchParam("pagelen", "30").
		Reply(200).
		Type("application/json").
		File("testdata/hooks.json")

	client, _ := New("https://api.bitbucket.org")
	got, _, err := client.Repositories.ListHooks(context.Background(), "atlassian/stash-example-plugin", scm.ListOptions{Size: 30, Page: 1})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Hook{}
	raw, _ := ioutil.ReadFile("testdata/hooks.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestRepositoryHookDelete(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Delete("/2.0/repositories/atlassian/stash-example-plugin/hooks/{d53603cc-3f67-45ea-b310-aaa5ef6ec061}").
		Reply(204).Done()

	client, _ := New("https://api.bitbucket.org")
	_, err := client.Repositories.DeleteHook(context.Background(), "atlassian/stash-example-plugin", "{d53603cc-3f67-45ea-b310-aaa5ef6ec061}")
	if err != nil {
		t.Error(err)
	}
}

func TestRepositoryHookCreate(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Post("/2.0/repositories/atlassian/stash-example-plugin/hooks").
		Reply(201).
		Type("application/json").
		File("testdata/hook.json")

	client, _ := New("https://api.bitbucket.org")
	got, _, err := client.Repositories.CreateHook(context.Background(), "atlassian/stash-example-plugin", &scm.HookInput{})
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Hook)
	raw, _ := ioutil.ReadFile("testdata/hook.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestConvertFromState(t *testing.T) {
	tests := []struct {
		src scm.State
		dst string
	}{
		{
			src: scm.StateCanceled,
			dst: "FAILED",
		},
		{
			src: scm.StateError,
			dst: "FAILED",
		},
		{
			src: scm.StateFailure,
			dst: "FAILED",
		},
		{
			src: scm.StatePending,
			dst: "INPROGRESS",
		},
		{
			src: scm.StateRunning,
			dst: "INPROGRESS",
		},
		{
			src: scm.StateSuccess,
			dst: "SUCCESSFUL",
		},
		{
			src: scm.StateUnknown,
			dst: "FAILED",
		},
	}
	for _, test := range tests {
		if got, want := convertFromState(test.src), test.dst; got != want {
			t.Errorf("Want state %v converted to %s", test.src, test.dst)
		}
	}
}

func TestConvertState(t *testing.T) {
	tests := []struct {
		src string
		dst scm.State
	}{
		{
			src: "FAILED",
			dst: scm.StateFailure,
		},
		{
			src: "INPROGRESS",
			dst: scm.StatePending,
		},
		{
			src: "SUCCESSFUL",
			dst: scm.StateSuccess,
		},
		{
			src: "STOPPED",
			dst: scm.StateUnknown,
		},
	}
	for _, test := range tests {
		if got, want := convertState(test.src), test.dst; got != want {
			t.Errorf("Want state %s converted to %v", test.src, test.dst)
		}
	}
}

func TestConvertPerms(t *testing.T) {
	tests := []struct {
		src *perm
		dst *scm.Perm
	}{
		{
			src: &perm{Permissions: "admin"},
			dst: &scm.Perm{Admin: true, Push: true, Pull: true},
		},
		{
			src: &perm{Permissions: "write"},
			dst: &scm.Perm{Admin: false, Push: true, Pull: true},
		},
		{
			src: &perm{Permissions: "read"},
			dst: &scm.Perm{Admin: false, Push: false, Pull: true},
		},
		{
			src: nil,
			dst: &scm.Perm{Admin: false, Push: false, Pull: false},
		},
	}
	for _, test := range tests {
		src := new(perms)
		if test.src != nil {
			src.Values = append(src.Values, test.src)
		}
		dst := convertPerms(src)
		if diff := cmp.Diff(test.dst, dst); diff != "" {
			t.Errorf("Unexpected Results")
			t.Log(diff)
		}
	}
}
