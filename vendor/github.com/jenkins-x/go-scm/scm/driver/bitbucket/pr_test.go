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

func TestPullFind(t *testing.T) {
	t.Skip()
}

func TestPullList(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Get("/2.0/repositories/octocat/hello-world/pullrequests").
		MatchParam("pagelen", "30").
		MatchParam("page", "1").
		MatchParam("state", "all").
		Reply(200).
		Type("application/json").
		//SetHeaders(mockHeaders).
		//SetHeaders(mockPageHeaders).
		File("testdata/pulls.json")

	client := NewDefault()
	got, _, err := client.PullRequests.List(context.Background(), "octocat/hello-world", scm.PullRequestListOptions{Page: 1, Size: 30, Open: true, Closed: true})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.PullRequest{}
	raw, _ := ioutil.ReadFile("testdata/pulls.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)

		data, _ := json.Marshal(got)
		t.Logf("got JSON: %s", data)
	}
}

func TestPullListChanges(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Get("/2.0/repositories/atlassian/atlaskit/pullrequests/1/diffstat").
		MatchParam("pagelen", "30").
		MatchParam("page", "1").
		Reply(200).
		Type("application/json").
		File("testdata/pr_diffstat.json")

	client, _ := New("https://api.bitbucket.org")
	got, _, err := client.PullRequests.ListChanges(context.Background(), "atlassian/atlaskit", 1, scm.ListOptions{Size: 30, Page: 1})
	if err != nil {
		t.Error(err)
	}

	want := []*scm.Change{}
	raw, _ := ioutil.ReadFile("testdata/pr_diffstat.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestPullMerge(t *testing.T) {
	t.Skip()
}

func TestPullClose(t *testing.T) {
	client, _ := New("https://api.bitbucket.org")
	_, err := client.PullRequests.Close(context.Background(), "atlassian/atlaskit", 1)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestPullCreate(t *testing.T) {
	client, _ := New("https://api.bitbucket.org")
	input := &scm.PullRequestInput{
		Title: "Bitbucket feature",
		Body:  "New Bitbucket feature",
		Head:  "new-feature",
		Base:  "master",
	}

	_, _, err := client.PullRequests.Create(context.Background(), "atlassian/atlaskit", input)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}
