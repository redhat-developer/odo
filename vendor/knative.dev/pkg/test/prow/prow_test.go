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

// prow_test.go contains unit tests for prow package

package prow

import (
	"os"
	"testing"
)

const (
	repoName       = "test-repo"
	invalidJobType = "invalid"
	testJobName    = "job_0"
)

var jobPathTests = []struct {
	in  string
	out string
}{
	{PeriodicJob, "logs/job_0"},
	{PostsubmitJob, "logs/job_0"},
	{PresubmitJob, "pr-logs/pull/knative_test-repo/0/job_0"},
	{BatchJob, "pr-logs/pull/batch/job_0"},
}

func TestJobPath(t *testing.T) {
	for _, testData := range jobPathTests {
		job := NewJob(testJobName, testData.in, repoName, 0)
		if job.StoragePath != testData.out {
			t.Errorf("Expected '%s', actual '%s'", testData.out, job.StoragePath)
		}
	}
}

func TestInvalidJobPath(t *testing.T) {
	oldLogFatalf := logFatalf
	defer func() { logFatalf = oldLogFatalf }()

	var exitString string
	expectedExitString := "unknown job spec type: invalid"
	logFatalf = func(string, ...interface{}) {
		exitString = expectedExitString
	}

	NewJob(testJobName, invalidJobType, repoName, 0)
	if exitString != expectedExitString {
		t.Fatalf("Expected: %s, actual: %s", exitString, expectedExitString)
	}
}

func TestGetArtifacts(t *testing.T) {
	dir := os.Getenv("ARTIFACTS")

	// Test we can read from the env var
	os.Setenv("ARTIFACTS", "test")
	v := GetLocalArtifactsDir()
	if v != "test" {
		t.Fatalf("Actual artifacts dir: '%s' and Expected: 'test'", v)
	}

	// Test we can use the default
	os.Setenv("ARTIFACTS", "")
	v = GetLocalArtifactsDir()
	if v != "artifacts" {
		t.Fatalf("Actual artifacts dir: '%s' and Expected: 'artifacts'", v)
	}

	// Set it to the original value
	os.Setenv("ARTIFACTS", dir)
}
