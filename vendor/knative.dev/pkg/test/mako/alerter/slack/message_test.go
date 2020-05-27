/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package slack

import (
	"os"
	"testing"
	"time"

	"knative.dev/pkg/test/mako/config"
	"knative.dev/pkg/test/slackutil/fakeslackutil"
)

var mh MessageHandler

func TestMain(m *testing.M) {
	client := fakeslackutil.NewFakeSlackClient()
	mh = MessageHandler{
		readClient:  client,
		writeClient: client,
		channels: []config.Channel{
			{Name: "test_channel1", Identity: "fsfdsf"},
			{Name: "test_channel2", Identity: "fdsfhfdh"},
		},
		dryrun: false,
	}
	os.Exit(m.Run())
}

func TestMessaging(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{"test name"},
		{`special name <>&'"`},
	}

	for _, tc := range testCases {
		historySizes := make([]int, len(mh.channels))
		for i, channel := range mh.channels {
			initHistory, err := mh.readClient.MessageHistory(channel.Identity, time.Now().Add(-1*time.Hour))
			if err != nil {
				t.Fatalf("expected to get the message history, but failed: %v", err)
			}
			historySizes[i] = len(initHistory)
		}

		firstMsg := "first message"
		if err := mh.SendAlert(tc.name, firstMsg); err != nil {
			t.Fatalf("expected to send the message, but failed: %v", err)
		}
		for i, channel := range mh.channels {
			history, err := mh.readClient.MessageHistory(channel.Identity, time.Now().Add(-1*time.Hour))
			if err != nil {
				t.Fatalf("expected to get the message history, but failed: %v", err)
			}
			if len(history) != historySizes[i]+1 {
				t.Fatalf("the message is expected to be successfully sent, but failed: %v", err)
			}
		}

		secondMsg := "second message"
		if err := mh.SendAlert(tc.name, secondMsg); err != nil {
			t.Fatalf("expected to send the message, but failed: %v", err)
		}
		for i, channel := range mh.channels {
			history, err := mh.readClient.MessageHistory(channel.Identity, time.Now().Add(-1*time.Hour))
			if err != nil {
				t.Fatalf("expected to get the message history, but failed: %v", err)
			}
			if len(history) != historySizes[i]+1 {
				t.Fatalf("the message history is expected to be unchanged, but now it's: %d", len(history))
			}
		}
	}
}

func TestDecoratedName(t *testing.T) {
	testCases := []struct {
		name           string
		expectedResult string
	}{
		{"demo test1", "[demo test1]"},
		{"[demo test2]", "[[demo test2]]"},
	}
	for _, tc := range testCases {
		actualResult := decoratedName(tc.name)
		if tc.expectedResult != actualResult {
			t.Fatalf("expected to get %q for decoratedName(%s), but got %q", tc.expectedResult, tc.name, actualResult)
		}
	}
}
