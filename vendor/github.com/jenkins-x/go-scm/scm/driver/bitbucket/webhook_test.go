// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitbucket

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
			event:  "repo:push",
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
			event:  "repo:push",
			before: "testdata/webhooks/push_tag_create.json",
			after:  "testdata/webhooks/push_tag_create.json.golden",
			obj:    new(scm.PushHook),
		},
		// delete
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "repo:push",
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
			event:  "repo:push",
			before: "testdata/webhooks/push_branch_create.json",
			after:  "testdata/webhooks/push_branch_create.json.golden",
			obj:    new(scm.PushHook),
		},
		// delete
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "repo:push",
			before: "testdata/webhooks/push_branch_delete.json",
			after:  "testdata/webhooks/push_branch_delete.json.golden",
			obj:    new(scm.BranchHook),
		},

		//
		// pull request events
		//

		// pull request created
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "pullrequest:created",
			before: "testdata/webhooks/pr_created.json",
			after:  "testdata/webhooks/pr_created.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request updated
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "pullrequest:updated",
			before: "testdata/webhooks/pr_updated.json",
			after:  "testdata/webhooks/pr_updated.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request fulfilled (merged)
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "pullrequest:fulfilled",
			before: "testdata/webhooks/pr_fulfilled.json",
			after:  "testdata/webhooks/pr_fulfilled.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request rejected (closed, declined)
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "pullrequest:rejected",
			before: "testdata/webhooks/pr_declined.json",
			after:  "testdata/webhooks/pr_declined.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// 		// pull request labeled
		// 		{
		// 			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
		// 			event:  "pull_request",
		// 			before: "samples/pr_labeled.json",
		// 			after:  "samples/pr_labeled.json.golden",
		// 			obj:    new(scm.PullRequestHook),
		// 		},
		// 		// pull request unlabeled
		// 		{
		// 			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
		// 			event:  "pull_request",
		// 			before: "samples/pr_unlabeled.json",
		// 			after:  "samples/pr_unlabeled.json.golden",
		// 			obj:    new(scm.PullRequestHook),
		// 		},
	}

	for _, test := range tests {
		before, err := ioutil.ReadFile(test.before)
		if err != nil {
			t.Error(err)
			continue
		}
		after, err := ioutil.ReadFile(test.after)
		if err != nil {
			t.Error(err)
			continue
		}

		buf := bytes.NewBuffer(before)
		r, _ := http.NewRequest("GET", "/?secret=71295b197fa25f4356d2fb9965df3f2379d903d7", buf)
		r.Header.Set("x-event-key", test.event)

		s := new(webhookService)
		o, err := s.Parse(r, secretFunc)
		if err != nil {
			t.Error(err)
			continue
		}

		err = json.Unmarshal(after, &test.obj)
		if err != nil {
			t.Error(err)
			continue
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
	}
}

func TestWebhookInvalid(t *testing.T) {
	f, _ := ioutil.ReadFile("testdata/webhooks/push.json")
	r, _ := http.NewRequest("GET", "/?secret=xxxxxinvalidxxxxxx", bytes.NewBuffer(f))
	r.Header.Set("x-event-key", "repo:push")

	s := new(webhookService)
	_, err := s.Parse(r, secretFunc)
	if err != scm.ErrSignatureInvalid {
		t.Errorf("Expect invalid signature error, got %v", err)
	}
}

func TestWebhookValidated(t *testing.T) {
	f, _ := ioutil.ReadFile("testdata/webhooks/push.json")
	r, _ := http.NewRequest("GET", "/?secret=71295b197fa25f4356d2fb9965df3f2379d903d7", bytes.NewBuffer(f))
	r.Header.Set("x-event-key", "repo:push")

	s := new(webhookService)
	_, err := s.Parse(r, secretFunc)
	if err != nil {
		t.Errorf("Expect valid signature, got %v", err)
	}
}

func secretFunc(scm.Webhook) (string, error) {
	return "71295b197fa25f4356d2fb9965df3f2379d903d7", nil
}
