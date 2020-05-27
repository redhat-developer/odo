// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stash

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jenkins-x/go-scm/scm"
)

func TestWebhooks(t *testing.T) {
	tests := []struct {
		sig    string
		event  string
		before string
		after  string
		obj    interface{}
	}{
		//
		// push events
		//

		// push hooks
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "repo:refs_changed",
			before: "testdata/webhooks/push.json",
			after:  "testdata/webhooks/push.json.golden",
			obj:    new(scm.PushHook),
		},

		//
		// tag events
		//

		// create
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "repo:refs_changed",
			before: "testdata/webhooks/push_tag_create.json",
			after:  "testdata/webhooks/push_tag_create.json.golden",
			obj:    new(scm.TagHook),
		},
		// delete
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "repo:refs_changed",
			before: "testdata/webhooks/push_tag_delete.json",
			after:  "testdata/webhooks/push_tag_delete.json.golden",
			obj:    new(scm.TagHook),
		},

		//
		// branch events
		//

		// create
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "repo:refs_changed",
			before: "testdata/webhooks/push_branch_create.json",
			after:  "testdata/webhooks/push_branch_create.json.golden",
			obj:    new(scm.BranchHook),
		},
		// delete
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "repo:refs_changed",
			before: "testdata/webhooks/push_branch_delete.json",
			after:  "testdata/webhooks/push_branch_delete.json.golden",
			obj:    new(scm.BranchHook),
		},

		//
		// pull request events
		//

		// pull request opened
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "pr:opened",
			before: "testdata/webhooks/pr_open.json",
			after:  "testdata/webhooks/pr_open.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request from ref updated
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "pr:from_ref_updated",
			before: "testdata/webhooks/pr_ref_updated.json",
			after:  "testdata/webhooks/pr_ref_updated.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request modified
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "pr:modified",
			before: "testdata/webhooks/pr_modified.json",
			after:  "testdata/webhooks/pr_modified.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request fulfilled (merged)
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "pr:merged",
			before: "testdata/webhooks/pr_merged.json",
			after:  "testdata/webhooks/pr_merged.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request rejected (closed, declined)
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "pr:declined",
			before: "testdata/webhooks/pr_declined.json",
			after:  "testdata/webhooks/pr_declined.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request comment
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "pr:comment:added",
			before: "testdata/webhooks/pr_comment.json",
			after:  "testdata/webhooks/pr_comment.json.golden",
			obj:    new(scm.PullRequestCommentHook),
		},
	}

	for _, test := range tests {
		if test.event != "pr:modified" {
			continue
		}
		t.Run(test.event, func(t *testing.T) {
			before, err := ioutil.ReadFile(test.before)
			if err != nil {
				t.Fatal(err)
			}
			after, err := ioutil.ReadFile(test.after)
			if err != nil {
				t.Fatal(err)
			}

			buf := bytes.NewBuffer(before)
			r, _ := http.NewRequest("GET", "/", buf)
			r.Header.Set("X-Event-Key", test.event)

			s := new(webhookService)
			o, err := s.Parse(r, secretFunc)
			if err != nil && err != scm.ErrSignatureInvalid {
				t.Fatal(err)
			}

			err = json.Unmarshal(after, &test.obj)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.obj, o); diff != "" {
				t.Errorf("Error unmarshaling %s", test.before)
				t.Log(diff)

				// debug only. remove once implemented
				json.NewEncoder(os.Stdout).Encode(o)
			}

			switch event := o.(type) {
			case *scm.PushHook:
				if !strings.HasPrefix(event.Ref, "refs/") {
					t.Errorf("Push hook reference must start with refs/")
				}
			case *scm.BranchHook:
				if strings.HasPrefix(event.Ref.Name, "refs/") {
					t.Errorf("Branch hook reference must not start with refs/")
				}
			case *scm.TagHook:
				if strings.HasPrefix(event.Ref.Name, "refs/") {
					t.Errorf("Branch hook reference must not start with refs/")
				}
			}
		})
	}
}

func TestWebhookInvalid(t *testing.T) {
	f, _ := ioutil.ReadFile("testdata/webhooks/push.json")
	r, _ := http.NewRequest("GET", "/", bytes.NewBuffer(f))
	r.Header.Set("X-Event-Key", "repo:refs_changed")
	r.Header.Set("X-Hub-Signature", "sha256=380f462cd2e160b84765144beabdad2e930a7ec5")

	s := new(webhookService)
	_, err := s.Parse(r, secretFunc)
	if err != scm.ErrSignatureInvalid {
		t.Errorf("Expect invalid signature error, got %v", err)
	}
}

func TestWebhookVerified(t *testing.T) {
	f, _ := ioutil.ReadFile("testdata/webhooks/push.json")
	r, _ := http.NewRequest("GET", "/", bytes.NewBuffer(f))
	r.Header.Set("X-Event-Key", "repo:refs_changed")
	r.Header.Set("X-Hub-Signature", "sha256=c90565fa018f3039414a7929c9187a147f1ac463076961c4cf411e3c67c541f8")

	s := new(webhookService)
	_, err := s.Parse(r, secretFunc)
	if err != nil {
		t.Errorf("Expect valid signature error, got %v", err)
	}
}

func secretFunc(scm.Webhook) (string, error) {
	return "71295b197fa25f4356d2fb9965df3f2379d903d7", nil
}
