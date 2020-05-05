// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gitea

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/jenkins-x/go-scm/scm"

	"github.com/google/go-cmp/cmp"
)

func TestWebhooks(t *testing.T) {
	tests := []struct {
		event  string
		before string
		after  string
		obj    interface{}
	}{
		// branch hooks
		{
			event:  "create",
			before: "testdata/webhooks/branch_create.json",
			after:  "testdata/webhooks/branch_create.json.golden",
			obj:    new(scm.BranchHook),
		},
		{
			event:  "delete",
			before: "testdata/webhooks/branch_delete.json",
			after:  "testdata/webhooks/branch_delete.json.golden",
			obj:    new(scm.BranchHook),
		},
		// tag hooks
		{
			event:  "create",
			before: "testdata/webhooks/tag_create.json",
			after:  "testdata/webhooks/tag_create.json.golden",
			obj:    new(scm.TagHook),
		},
		{
			event:  "delete",
			before: "testdata/webhooks/tag_delete.json",
			after:  "testdata/webhooks/tag_delete.json.golden",
			obj:    new(scm.TagHook),
		},
		// push hooks
		{
			event:  "push",
			before: "testdata/webhooks/push.json",
			after:  "testdata/webhooks/push.json.golden",
			obj:    new(scm.PushHook),
		},
		// issue hooks
		{
			event:  "issues",
			before: "testdata/webhooks/issues_opened.json",
			after:  "testdata/webhooks/issues_opened.json.golden",
			obj:    new(scm.IssueHook),
		},
		// issue comment hooks
		{
			event:  "issue_comment",
			before: "testdata/webhooks/issue_comment_created.json",
			after:  "testdata/webhooks/issue_comment_created.json.golden",
			obj:    new(scm.IssueCommentHook),
		},
		// pull request hooks
		{
			event:  "pull_request",
			before: "testdata/webhooks/pull_request_opened.json",
			after:  "testdata/webhooks/pull_request_opened.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		{
			event:  "pull_request",
			before: "testdata/webhooks/pull_request_edited.json",
			after:  "testdata/webhooks/pull_request_edited.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		{
			event:  "pull_request",
			before: "testdata/webhooks/pull_request_synchronized.json",
			after:  "testdata/webhooks/pull_request_synchronized.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		{
			event:  "pull_request",
			before: "testdata/webhooks/pull_request_closed.json",
			after:  "testdata/webhooks/pull_request_closed.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		{
			event:  "pull_request",
			before: "testdata/webhooks/pull_request_reopened.json",
			after:  "testdata/webhooks/pull_request_reopened.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		{
			event:  "pull_request",
			before: "testdata/webhooks/pull_request_merged.json",
			after:  "testdata/webhooks/pull_request_merged.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request comment hooks
		{
			event:  "issue_comment",
			before: "testdata/webhooks/pull_request_comment_created.json",
			after:  "testdata/webhooks/pull_request_comment_created.json.golden",
			obj:    new(scm.PullRequestCommentHook),
		},
	}

	for _, test := range tests {
		t.Run(test.before, func(t *testing.T) {
			t.Run(test.before, func(t *testing.T) {
				before, err := ioutil.ReadFile(test.before)
				if err != nil {
					t.Error(err)
					return
				}
				after, err := ioutil.ReadFile(test.after)
				if err != nil {
					t.Error(err)
					return
				}

				buf := bytes.NewBuffer(before)
				r, _ := http.NewRequest("GET", "/", buf)
				r.Header.Set("X-Gitea-Event", test.event)
				r.Header.Set("X-Gitea-Delivery", "ee8d97b4-1479-43f1-9cac-fbbd1b80da55")

				s := new(webhookService)
				o, err := s.Parse(r, secretFunc)
				if err != nil && err != scm.ErrSignatureInvalid {
					t.Error(err)
					return
				}

				err = json.Unmarshal(after, &test.obj)
				if err != nil {
					t.Error(err)
					return
				}

				if diff := cmp.Diff(test.obj, o); diff != "" {
					t.Errorf("Error unmarshaling %s", test.before)
					t.Log(diff)

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
		})
	}
}

func TestWebhook_ErrUnknownEvent(t *testing.T) {
	f, _ := ioutil.ReadFile("testdata/webhooks/pull_request_edited.json")
	r, _ := http.NewRequest("GET", "/", bytes.NewBuffer(f))

	s := new(webhookService)
	_, err := s.Parse(r, secretFunc)
	if !scm.IsUnknownWebhook(err) {
		t.Errorf("Expect unknown event error, got %v", err)
	}
}

func TestWebhookInvalid(t *testing.T) {
	f, _ := ioutil.ReadFile("testdata/webhooks/pull_request_edited.json")
	r, _ := http.NewRequest("GET", "/?secert=xxxxxxinvalidxxxxx", bytes.NewBuffer(f))
	r.Header.Set("X-Gitea-Event", "pull_request")
	r.Header.Set("X-Gitea-Delivery", "ee8d97b4-1479-43f1-9cac-fbbd1b80da55")
	r.Header.Set("X-Gitea-Signature", "failfailfailfail")

	s := new(webhookService)
	_, err := s.Parse(r, secretFunc)
	if err != scm.ErrSignatureInvalid {
		t.Errorf("Expect invalid signature error, got %v", err)
	}
}

func TestWebhook_Validated(t *testing.T) {
	f, _ := ioutil.ReadFile("testdata/webhooks/pull_request_edited.json")
	r, _ := http.NewRequest("GET", "/?secret=71295b197fa25f4356d2fb9965df3f2379d903d7", bytes.NewBuffer(f))
	r.Header.Set("X-Gitea-Event", "pull_request")
	r.Header.Set("X-Gitea-Delivery", "ee8d97b4-1479-43f1-9cac-fbbd1b80da55")
	r.Header.Set("X-Gitea-Signature", "a31111f057bafe895837f4a93c0f1f528919c199a20438b1fc8e23485780a33a")

	s := new(webhookService)
	_, err := s.Parse(r, secretFunc)
	if err != nil {
		t.Errorf("Expect valid signature, got %v", err)
	}
}

func TestWebhook_MissingSignature(t *testing.T) {
	f, _ := ioutil.ReadFile("testdata/webhooks/pull_request_edited.json")
	r, _ := http.NewRequest("GET", "/", bytes.NewBuffer(f))
	r.Header.Set("X-Gitea-Event", "pull_request")
	r.Header.Set("X-Gitea-Delivery", "ee8d97b4-1479-43f1-9cac-fbbd1b80da55")

	s := new(webhookService)
	_, err := s.Parse(r, secretFunc)
	if err != scm.ErrSignatureInvalid {
		t.Errorf("Expect invalid signature error, got %v", err)
	}
}

func secretFunc(scm.Webhook) (string, error) {
	return "71295b197fa25f4356d2fb9965df3f2379d903d7", nil
}
