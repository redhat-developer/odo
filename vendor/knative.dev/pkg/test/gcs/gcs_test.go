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

package gcs

import (
	"errors"
	"reflect"
	"testing"
)

func TestGetConsoleURL(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{
			name: "Missing protocol",
			arg:  "knative-prow/logs/ci-knative-docs-continuous/1132539579983728640/",
			want: "https://console.cloud.google.com/storage/browser/knative-prow/logs/ci-knative-docs-continuous/1132539579983728640",
		},
		{
			name: "gs protocol",
			arg:  "gs://knative-prow/logs/ci-knative-client-go-coverage/1139250680293232640",
			want: "https://console.cloud.google.com/storage/browser/knative-prow/logs/ci-knative-client-go-coverage/1139250680293232640",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := GetConsoleURL(tt.arg); got != tt.want {
				t.Errorf("GetConsoleURL(%v), got: %v, want: %v", tt.arg, got, tt.want)
			}
		})
	}
}

func TestBuildLogPath(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{
			name: "Trailing slash",
			arg:  "gs://knative-prow/logs/ci-knative-client-go-coverage/1139250680293232640/",
			want: "gs://knative-prow/logs/ci-knative-client-go-coverage/1139250680293232640/build-log.txt",
		},
		{
			name: "No Trailing slash",
			arg:  "gs://knative-prow/logs/ci-knative-client-go-coverage/1139250680293232640",
			want: "gs://knative-prow/logs/ci-knative-client-go-coverage/1139250680293232640/build-log.txt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := BuildLogPath(tt.arg); got != tt.want {
				t.Errorf("BuildLogPath(%v), got: %v, want: %v", tt.arg, got, tt.want)
			}
		})
	}
}

func TestLinkToBucketAndObject(t *testing.T) {
	type result struct {
		bucket string
		object string
	}

	tests := []struct {
		name string
		arg  string
		want *result
		err  error
	}{
		{
			name: "Valid gcsUrl",
			arg:  "gs://knative-prow/logs/ci-knative-client-go-coverage/1139250680293232640/build-log.txt",
			want: &result{
				bucket: "knative-prow",
				object: "logs/ci-knative-client-go-coverage/1139250680293232640/build-log.txt",
			},
			err: nil,
		},
		{
			name: "Invalid gcsUrl - No slash",
			arg:  "knative-prow-no-object",
			err:  errors.New("the gsUrl (knative-prow-no-object) cannot be converted to bucket/object"),
		},
		{
			name: "Invalid gcsUrl - No object",
			arg:  "knative-prow/",
			err:  errors.New("the gsUrl (knative-prow/) cannot be converted to bucket/object"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, o, err := linkToBucketAndObject(tt.arg)
			got := &result{b, o}
			if tt.err != nil && tt.err.Error() != err.Error() {
				t.Errorf("linktoBucketAndObject(%v), got error: %v, want error: %v", tt.arg, err, tt.err)
			}
			if tt.err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("linktoBucketAndObject(%v), got: %v, want: %v", tt.arg, got, tt.want)
			}
		})
	}
}
