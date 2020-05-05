// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stash

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

	gock.New("http://example.com:7990").
		Get("/rest/api/1.0/projects/PRJ/repos/my-repo").
		Reply(200).
		Type("application/json").
		File("testdata/repo.json")

	client, _ := New("http://example.com:7990")
	got, _, err := client.Repositories.Find(context.Background(), "PRJ/my-repo")
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

	gock.New("http://example.com:7990").
		Get("/rest/api/1.0/projects/dev/repos/null").
		Reply(404).
		Type("application/json").
		File("testdata/error.json")

	client, _ := New("http://example.com:7990")
	_, _, err := client.Repositories.Find(context.Background(), "dev/null")
	if err == nil {
		t.Errorf("Expect not found message")
	}

	if got, want := err.Error(), "Project dev does not exist."; got != want {
		t.Errorf("Want error message %q, got %q", want, got)
	}
}

func TestRepositoryPerms(t *testing.T) {
	defer gock.Off()

	gock.New("http://example.com:7990").
		Get("/rest/api/1.0/projects/PRJ/repos/my-repo").
		Reply(200).
		Type("application/json").
		File("testdata/repo.json")

	gock.New("http://example.com:7990").
		Get("/rest/api/1.0/projects/PRJ/repos/my-repo/webhooks").
		Reply(200).
		Type("application/json").
		File("testdata/webhooks.json")

	client, _ := New("http://example.com:7990")
	got, _, err := client.Repositories.FindPerms(context.Background(), "PRJ/my-repo")
	if err != nil {
		t.Error(err)
	}

	want := &scm.Perm{
		Pull:  true,
		Push:  true,
		Admin: true,
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	if gock.IsPending() {
		t.Errorf("pending API requests")
	}
}

func TestRepositoryPerms_ReadOnly(t *testing.T) {
	defer gock.Off()

	gock.New("http://example.com:7990").
		Get("/rest/api/1.0/projects/PRJ/repos/my-repo").
		Reply(200).
		Type("application/json").
		File("testdata/repo.json")

	gock.New("http://example.com:7990").
		Get("/rest/api/1.0/projects/PRJ/repos/my-repo/webhooks").
		Reply(404).
		Type("application/json")

	gock.New("http://example.com:7990").
		Get("/rest/api/1.0/repos").
		Reply(404).
		Type("application/json").
		File("testdata/repo.json")

	client, _ := New("http://example.com:7990")
	got, _, err := client.Repositories.FindPerms(context.Background(), "PRJ/my-repo")
	if err != nil {
		t.Error(err)
	}

	want := &scm.Perm{
		Pull:  true,
		Push:  false,
		Admin: false,
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	if gock.IsPending() {
		t.Errorf("pending API requests")
	}
}

func TestRepositoryPerms_Write(t *testing.T) {
	defer gock.Off()

	gock.New("http://example.com:7990").
		Get("/rest/api/1.0/projects/PRJ/repos/my-repo").
		Reply(200).
		Type("application/json").
		File("testdata/repo.json")

	gock.New("http://example.com:7990").
		Get("/rest/api/1.0/projects/PRJ/repos/my-repo/webhooks").
		Reply(404).
		Type("application/json")

	gock.New("http://example.com:7990").
		Get("/rest/api/1.0/repos").
		MatchParam("size", "1000").
		MatchParam("permission", "REPO_WRITE").
		MatchParam("project", "PRJ").
		MatchParam("name", "my-repo").
		Reply(200).
		Type("application/json").
		File("testdata/repos.json")

	client, _ := New("http://example.com:7990")
	got, _, err := client.Repositories.FindPerms(context.Background(), "PRJ/my-repo")
	if err != nil {
		t.Error(err)
	}

	want := &scm.Perm{
		Pull:  true,
		Push:  true,
		Admin: false,
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	if gock.IsPending() {
		t.Errorf("pending API requests")
	}
}

func TestRepositoryPerms_Forbidden(t *testing.T) {
	defer gock.Off()

	gock.New("http://example.com:7990").
		Get("/rest/api/1.0/projects/PRJ/repos/my-repo").
		Reply(404).
		Type("application/json")

	client, _ := New("http://example.com:7990")
	got, _, err := client.Repositories.FindPerms(context.Background(), "PRJ/my-repo")
	if err != nil {
		t.Error(err)
	}

	want := &scm.Perm{
		Pull:  false,
		Push:  false,
		Admin: false,
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestRepositoryList(t *testing.T) {
	defer gock.Off()

	gock.New("http://example.com:7990").
		Get("/rest/api/1.0/repos").
		MatchParam("limit", "25").
		MatchParam("start", "50").
		MatchParam("permission", "REPO_READ").
		Reply(200).
		Type("application/json").
		File("testdata/repos.json")

	client, _ := New("http://example.com:7990")
	got, res, err := client.Repositories.List(context.Background(), scm.ListOptions{Page: 3, Size: 25})
	if err != nil {
		t.Error(err)
	}

	if got, want := res.Page.First, 1; got != want {
		t.Errorf("Want Page.First %d, got %d", want, got)
	}
	if got, want := res.Page.Next, 4; got != want {
		t.Errorf("Want Page.Next %d, got %d", want, got)
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

	gock.New("http://example.com:7990").
		Get("/rest/build-status/1.0/commits/b02e90353e4c94cda868dbcdb2301c5691a78b6c").
		Reply(200).
		Type("application/json").
		File("testdata/commit_build_status.json")

	client, _ := New("http://example.com:7990")

	got, _, err := client.Repositories.ListStatus(context.Background(), "", "b02e90353e4c94cda868dbcdb2301c5691a78b6c", scm.ListOptions{Size: 30, Page: 1})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Status{}
	raw, _ := ioutil.ReadFile("testdata/commit_build_status.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestStatusCreate(t *testing.T) {
	defer gock.Off()

	gock.New("http://example.com:7990").
		Post("/rest/build-status/1.0/commits/a6e5e7d797edf751cbd839d6bd4aef86c941eec9").
		Reply(204)

	in := &scm.StatusInput{
		Desc:   "Build has completed successfully",
		Label:  "continuous-integration/drone/pull",
		State:  scm.StateSuccess,
		Target: "https://ci.example.com/1000/output",
	}

	client, _ := New("http://example.com:7990")
	_, _, err := client.Repositories.CreateStatus(context.Background(), "PRJ/my-repo", "a6e5e7d797edf751cbd839d6bd4aef86c941eec9", in)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestRepositoryHookFind(t *testing.T) {
	defer gock.Off()

	gock.New("http://example.com:7990").
		Get("/rest/api/1.0/projects/PRJ/repos/my-repo/webhooks/1").
		Reply(200).
		Type("application/json").
		File("testdata/webhook.json")

	client, _ := New("http://example.com:7990")
	got, _, err := client.Repositories.FindHook(context.Background(), "PRJ/my-repo", "1")
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Hook)
	raw, _ := ioutil.ReadFile("testdata/webhook.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestRepositoryHookList(t *testing.T) {
	defer gock.Off()

	gock.New("http://example.com:7990").
		Get("/rest/api/1.0/projects/PRJ/repos/my-repo/webhooks").
		MatchParam("limit", "30").
		Reply(200).
		Type("application/json").
		File("testdata/webhooks.json")

	client, _ := New("http://example.com:7990")
	got, _, err := client.Repositories.ListHooks(context.Background(), "PRJ/my-repo", scm.ListOptions{Size: 30, Page: 1})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Hook{}
	raw, _ := ioutil.ReadFile("testdata/webhooks.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestRepositoryHookDelete(t *testing.T) {
	defer gock.Off()

	gock.New("http://example.com:7990").
		Delete("/rest/api/1.0/projects/PRJ/repos/my-repo/webhooks/1").
		Reply(200).
		Type("application/json")

	client, _ := New("http://example.com:7990")
	_, err := client.Repositories.DeleteHook(context.Background(), "PRJ/my-repo", "1")
	if err != nil {
		t.Error(err)
	}
}

func TestRepositoryHookCreate(t *testing.T) {
	defer gock.Off()

	gock.New("http://example.com:7990").
		Post("/rest/api/1.0/projects/PRJ/repos/my-repo/webhooks").
		Reply(201).
		Type("application/json").
		File("testdata/webhook.json")

	client, _ := New("http://example.com:7990")
	got, _, err := client.Repositories.CreateHook(context.Background(), "PRJ/my-repo", &scm.HookInput{
		Name:   "example",
		Target: "http://example.com",
		Secret: "12345",
		Events: scm.HookEvents{
			Branch:             true,
			PullRequest:        true,
			PullRequestComment: true,
			Push:               true,
			Tag:                true,
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Hook)
	raw, _ := ioutil.ReadFile("testdata/webhook.json.golden")
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
