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

func TestContentFind(t *testing.T) {
	defer gock.Off()

	gock.New("http://example.com:7990").
		Get("/rest/api/1.0/projects/PRJ/repos/my-repo/raw/README").
		MatchParam("at", "5c64a07cd6c0f21b753bf261ef059c7e7633c50a").
		Reply(200).
		Type("text/plain").
		File("testdata/content.txt")

	client, _ := New("http://example.com:7990")
	got, _, err := client.Contents.Find(context.Background(), "PRJ/my-repo", "README", "5c64a07cd6c0f21b753bf261ef059c7e7633c50a")
	if err != nil {
		t.Error(err)
	}

	want := new(scm.Content)
	raw, _ := ioutil.ReadFile("testdata/content.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestContentCreate(t *testing.T) {
	content := new(contentService)
	_, err := content.Create(context.Background(), "atlassian/atlaskit", "README", nil)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestContentUpdate(t *testing.T) {
	content := new(contentService)
	_, err := content.Update(context.Background(), "atlassian/atlaskit", "README", nil)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestContentDelete(t *testing.T) {
	content := new(contentService)
	_, err := content.Delete(context.Background(), "atlassian/atlaskit", "README", "master")
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}
