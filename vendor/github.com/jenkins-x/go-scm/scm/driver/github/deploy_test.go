// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/h2non/gock"
	"github.com/jenkins-x/go-scm/scm"
)

func TestDeploymentFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/repos/octocat/example/deployments/1").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/deploy.json")

	client := NewDefault()
	got, res, err := client.Deployments.Find(context.Background(), "octocat/example", "1")
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Deployment)
	raw, _ := ioutil.ReadFile("testdata/deploy.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)

		logGot(t, got)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestDeploymentNotFound(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/repos/dev/null/deployments/999").
		Reply(404).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/error.json")

	client := NewDefault()
	_, _, err := client.Deployments.Find(context.Background(), "dev/null", "999")
	if err == nil {
		t.Errorf("Expect Not Found error")
		return
	}
	if got, want := err.Error(), "Not Found"; got != want {
		t.Errorf("Want error %q, got %q", want, got)
	}
}

func TestDeploymentList(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/repos/octocat/example/deployments").
		MatchParam("page", "1").
		MatchParam("per_page", "30").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/deploys.json")

	client := NewDefault()
	got, res, err := client.Deployments.List(context.Background(), "octocat/example", scm.ListOptions{Page: 1, Size: 30})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Deployment{}
	raw, _ := ioutil.ReadFile("testdata/deploys.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)

		logGot(t, got)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
	t.Run("Page", testPage(res))
}

func TestDeploymentCreate(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Post("/repos/octocat/example/deployments").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/deploy_create.json")

	in := &scm.DeploymentInput{}

	client := NewDefault()
	got, res, err := client.Deployments.Create(context.Background(), "octocat/example", in)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Deployment)
	raw, _ := ioutil.ReadFile("testdata/deploy_create.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)

		logGot(t, got)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestDeploymentStatusList(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/repos/octocat/example/deployments/1/statuses").
		MatchParam("page", "1").
		MatchParam("per_page", "30").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/deploy_statuses.json")

	client := NewDefault()
	got, res, err := client.Deployments.ListStatus(context.Background(), "octocat/example", "1", scm.ListOptions{Page: 1, Size: 30})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.DeploymentStatus{}
	raw, _ := ioutil.ReadFile("testdata/deploy_statuses.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)

		logGot(t, got)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
	t.Run("Page", testPage(res))
}

func TestDeploymentStatusFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/repos/octocat/example/deployments/1/statuses/1").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/deploy_status.json")

	client := NewDefault()
	got, res, err := client.Deployments.FindStatus(context.Background(), "octocat/example", "1", "1")
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.DeploymentStatus)
	raw, _ := ioutil.ReadFile("testdata/deploy_status.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)

		logGot(t, got)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestDeploymentStatusCreate(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Post("repos/octocat/example/deployments/1/statuses").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/deploy_status_create.json")

	in := &scm.DeploymentStatusInput{}

	client := NewDefault()
	got, res, err := client.Deployments.CreateStatus(context.Background(), "octocat/example", "1", in)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.DeploymentStatus)
	raw, _ := ioutil.ReadFile("testdata/deploy_status_create.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)

		logGot(t, got)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func logGot(t *testing.T, got interface{}) {
	data, _ := json.Marshal(got)
	t.Log("got JSON:")
	t.Log(string(data))
}
