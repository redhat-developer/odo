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

package config

import "testing"

func TestNonDefaultSlackChannelsConfig(t *testing.T) {
	configStr := `
benchmarkChannels:
  benchmark1:
  - name: channel11
    identity: identity11
  - name: channel12
    identity: identity12
  - name: channel13
    identity: identity13
  benchmark2:
  - name: channel21
    identity: identity21
  - name: channel22
    identity: identity22`

	testCases := []struct {
		benchmarkName        string
		expectedChannelCount int
	}{
		{"benchmark1", 3},
		{"benchmark2", 2},
	}
	for _, v := range testCases {
		channels := getSlackChannels(configStr, v.benchmarkName)
		if v.expectedChannelCount != len(channels) {
			t.Fatalf("expected to get %q channels for benchmark %q but actual is %q", v.expectedChannelCount, v.benchmarkName, len(channels))
		}
	}
}

func TestDefaultSlackChannelsConfig(t *testing.T) {
	configStr := `
benchmarkChannels:
  benchmark1:
  - name: channel11
    identity: identity11
  - name: channel12
    identity: identity12`

	channels := getSlackChannels(configStr, "non-existing-benchmark-name")
	if len(channels) != 1 {
		t.Fatalf("expected to get one default channel but actual is %q", len(channels))
	}
	if channels[0].Name != defaultChannel.Name || channels[0].Identity != defaultChannel.Identity {
		t.Fatalf("expected to get the default channel but actual is %q", channels[0])
	}
}

func TestEmptySlackChannelsConfig(t *testing.T) {
	configStr := ""
	channels := getSlackChannels(configStr, "channel-name")
	if len(channels) != 1 {
		t.Fatalf("expected to get one default channel but actual is %q", len(channels))
	}
	if channels[0].Name != defaultChannel.Name || channels[0].Identity != defaultChannel.Identity {
		t.Fatalf("expected to get the default channel but actual is %q", channels[0])
	}
}
